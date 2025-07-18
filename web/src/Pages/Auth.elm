module Pages.Auth exposing (Banner, FormVariant, Model, Msg, page)

import Api
import Api.Auth
import Auth.User
import Components.Box
import Components.Form
import Components.Utils
import Data.Credentials exposing (Credentials)
import Dict
import Effect exposing (Effect)
import Html as H exposing (Html)
import Html.Attributes as A
import Html.Events as E
import Layouts
import Page exposing (Page)
import Route exposing (Route)
import Route.Path
import Shared
import Time exposing (Posix)
import View exposing (View)


page : Shared.Model -> Route () -> Page Model Msg
page shared route =
    Page.new
        { init = init shared route
        , update = update
        , subscriptions = subscriptions
        , view = view
        }
        |> Page.withLayout (\_ -> Layouts.Header {})



-- INIT


type alias Model =
    { email : String
    , password : String
    , passwordAgain : String
    , isSubmittingForm : Bool
    , banner : Banner
    , formVariant : FormVariant
    , lastClicked : Maybe Posix
    , now : Maybe Posix
    }


init : Shared.Model -> Route () -> () -> ( Model, Effect Msg )
init shared route () =
    let
        formVariant =
            case Dict.get "token" route.query of
                Just token ->
                    SetNewPassword token

                Nothing ->
                    SignIn
    in
    ( { formVariant = formVariant
      , isSubmittingForm = False
      , email = ""
      , password = ""
      , passwordAgain = ""
      , lastClicked = Nothing
      , banner = Hidden
      , now = Nothing
      }
    , case shared.user of
        Auth.User.SignedIn _ ->
            Effect.pushRoutePath Route.Path.Home_

        _ ->
            Effect.none
    )



-- UPDATE


type Msg
    = Tick Posix
    | UserUpdatedInput Field String
    | UserChangedFormVariant FormVariant
    | UserClickedSubmit
    | UserClickedResendActivationEmail
    | ApiSignInResponded (Result Api.Error Credentials)
    | ApiSignUpResponded (Result Api.Error ())
    | ApiForgotPasswordResponded (Result Api.Error ())
    | ApiSetNewPasswordResponded (Result Api.Error ())
    | ApiResendVerificationEmail (Result Api.Error ())


type Field
    = Email
    | Password
    | PasswordAgain


type alias ResetPasswordToken =
    String


type FormVariant
    = SignIn
    | SignUp
    | ForgotPassword
    | SetNewPassword ResetPasswordToken


type Banner
    = Hidden
    | ResendVerificationEmail
    | Error Api.Error
    | CheckEmail


update : Msg -> Model -> ( Model, Effect Msg )
update msg model =
    case msg of
        Tick now ->
            ( { model | now = Just now }, Effect.none )

        UserClickedSubmit ->
            ( { model | isSubmittingForm = True }
            , case model.formVariant of
                SignIn ->
                    Api.Auth.signin
                        { onResponse = ApiSignInResponded
                        , email = model.email
                        , password = model.password
                        }

                SignUp ->
                    Api.Auth.signup
                        { onResponse = ApiSignUpResponded
                        , email = model.email
                        , password = model.password
                        }

                ForgotPassword ->
                    Api.Auth.forgotPassword { onResponse = ApiForgotPasswordResponded, email = model.email }

                SetNewPassword token ->
                    Api.Auth.resetPassword { onResponse = ApiSetNewPasswordResponded, token = token, password = model.password }
            )

        UserClickedResendActivationEmail ->
            ( { model | lastClicked = model.now }
            , Api.Auth.resendVerificationEmail
                { onResponse = ApiResendVerificationEmail
                , email = model.email
                }
            )

        UserChangedFormVariant variant ->
            ( { model | formVariant = variant }, Effect.none )

        UserUpdatedInput Email email ->
            ( { model | email = email }, Effect.none )

        UserUpdatedInput Password password ->
            ( { model | password = password }, Effect.none )

        UserUpdatedInput PasswordAgain passwordAgain ->
            ( { model | passwordAgain = passwordAgain }, Effect.none )

        ApiSignInResponded (Ok credentials) ->
            ( { model | isSubmittingForm = False }, Effect.signin credentials )

        ApiSignInResponded (Err err) ->
            if Api.isNotVerified err then
                ( { model | isSubmittingForm = False, banner = ResendVerificationEmail }, Effect.none )

            else
                ( { model | isSubmittingForm = False, banner = Error err }, Effect.none )

        ApiSignUpResponded (Ok ()) ->
            ( { model | isSubmittingForm = False, banner = ResendVerificationEmail }, Effect.none )

        ApiSignUpResponded (Err err) ->
            ( { model | isSubmittingForm = False, banner = Error err }, Effect.none )

        ApiResendVerificationEmail (Ok ()) ->
            ( model, Effect.none )

        ApiResendVerificationEmail (Err err) ->
            ( { model | banner = Error err }, Effect.none )

        ApiSetNewPasswordResponded (Ok ()) ->
            ( { model | isSubmittingForm = False, formVariant = SignIn, password = "", passwordAgain = "" }, Effect.replaceRoutePath Route.Path.Auth )

        ApiSetNewPasswordResponded (Err err) ->
            ( { model | isSubmittingForm = False, banner = Error err }, Effect.none )

        ApiForgotPasswordResponded (Ok ()) ->
            ( { model | isSubmittingForm = False, banner = CheckEmail }, Effect.none )

        ApiForgotPasswordResponded (Err err) ->
            ( { model | isSubmittingForm = False, banner = Error err }, Effect.none )


subscriptions : Model -> Sub Msg
subscriptions model =
    if model.banner == ResendVerificationEmail then
        Time.every 1000 Tick

    else
        Sub.none



-- VIEW


view : Model -> View Msg
view model =
    { title = "Authentication"
    , body =
        [ H.div [ A.class "min-h-screen flex items-center justify-center bg-gray-50" ]
            [ Components.Utils.roundedBoxContainer
                -- TODO: add oauth buttons
                [ case ( model.apiError, model.showVerifyBanner ) of
                    ( Just error, False ) ->
                        Components.Error.error (Api.errorMessage error)

                    ( Nothing, True ) ->
                        viewVerificationBanner model.now model.lastClicked

                    _ ->
                        H.text ""
                , viewBoxHeader model.formVariant
                , H.div [ A.class "px-6 pb-6 space-y-4" ]
                    [ viewChangeVariant model.formVariant
                    , H.div [ A.class "border-t border-gray-200" ] []
                    , viewForm model
                    ]
                ]
            ]
        ]
    }


viewBanner : Model -> Html Msg
viewBanner model =
    case model.banner of
        Hidden ->
            H.text ""

        Error err ->
            Components.Box.error (Api.errorMessage err)

        CheckEmail ->
            Components.Box.success { header = "Check your email!", body = "To continue with resetting your password please check the email we've sent." }

        ResendVerificationEmail ->
            viewVerificationBanner model.now model.lastClicked


viewVerificationBanner : Maybe Posix -> Maybe Posix -> Html Msg
viewVerificationBanner now lastClicked =
    let
        buttonClassesBase =
            "w-full px-4 py-2 rounded-md focus:outline-none focus:ring-2 focus:ring-black focus:ring-offset-2 transition-colors mt-3"

        buttonClasses active =
            if active then
                buttonClassesBase ++ " border border-gray-300 text-gray-700 hover:bg-gray-50"

            else
                buttonClassesBase ++ " border border-gray-300 text-gray-400 cursor-not-allowed"

        timeLeftSeconds =
            case ( now, lastClicked ) of
                ( Just now_, Just last ) ->
                    let
                        elapsedMs =
                            Time.posixToMillis now_ - Time.posixToMillis last
                    in
                    max 0 ((30 * 1000 - elapsedMs) // 1000)

                _ ->
                    0

        canClick : Bool
        canClick =
            timeLeftSeconds == 0
    in
    Components.Box.successBox
        [ H.div [ A.class "font-medium text-green-800 mb-2" ] [ H.text "Check your email!" ]
        , H.p [ A.class "text-green-800 text-sm" ] [ H.text "Please verify your account to continue. We've sent a verification link to your email — click it to activate your account." ]
        , H.button
            [ A.class (buttonClasses canClick)
            , E.onClick UserClickedResendActivationEmail
            , A.disabled (not canClick)
            ]
            [ H.text "Resend verification email" ]
        , Components.Utils.viewIf (not canClick)
            (H.p
                [ A.class "text-gray-600 text-xs mt-2" ]
                [ H.text ("You can request a new verification email in " ++ String.fromInt timeLeftSeconds ++ " seconds.") ]
            )
        ]


viewBoxHeader : FormVariant -> Html Msg
viewBoxHeader variant =
    let
        ( title, description ) =
            case variant of
                SignIn ->
                    ( "Welcome Back", "Enter your credentials to access your account" )

                SignUp ->
                    ( "Create Account", "Enter your information to create your account" )

                ForgotPassword ->
                    ( "Forgot Password", "Enter your email to reset your password" )

                SetNewPassword _ ->
                    ( "Set New Password", "Enter your new password to reset your account" )
    in
    H.div [ A.class "p-6 pb-4" ]
        [ H.h1 [ A.class "text-2xl font-bold text-center mb-2" ] [ H.text title ]
        , H.p [ A.class "text-center text-gray-600 text-sm" ] [ H.text description ]
        ]


viewChangeVariant : FormVariant -> Html Msg
viewChangeVariant variant =
    let
        buttonClasses active =
            let
                base =
                    "flex-1 px-4 py-2 rounded-md font-medium transition-colors"
            in
            if active then
                base ++ " bg-black text-white"

            else
                base ++ " bg-white text-black border border-gray-300 hover:bg-gray-50"
    in
    H.div [ A.class "flex gap-2" ]
        [ H.button
            [ A.class (buttonClasses (variant == SignIn))
            , A.disabled (variant == SignIn)
            , E.onClick (UserChangedFormVariant SignIn)
            ]
            [ H.text "Sign In" ]
        , H.button
            [ A.class (buttonClasses (variant == SignUp))
            , A.disabled (variant == SignUp)
            , E.onClick (UserChangedFormVariant SignUp)
            ]
            [ H.text "Sign Up" ]
        ]


viewForm : Model -> Html Msg
viewForm model =
    H.form
        [ A.class "space-y-4"
        , E.onSubmit UserClickedSubmit
        ]
        (case model.formVariant of
            SignIn ->
                [ viewFormInput { field = Email, value = model.email }
                , viewFormInput { field = Password, value = model.password }
                , viewForgotPassword
                , viewSubmitButton model
                ]

            SignUp ->
                [ viewFormInput { field = Email, value = model.email }
                , viewFormInput { field = Password, value = model.password }
                , viewFormInput { field = PasswordAgain, value = model.passwordAgain }
                , viewSubmitButton model
                ]

            ForgotPassword ->
                [ viewFormInput { field = Email, value = model.email }
                , viewSubmitButton model
                ]

            SetNewPassword token ->
                [ viewFormInput { field = Password, value = model.password }
                , viewFormInput { field = PasswordAgain, value = model.passwordAgain }
                , H.input [ A.type_ "hidden", A.value token, A.name "token" ] []
                , viewSubmitButton model
                ]
        )


viewFormInput : { field : Field, value : String } -> Html Msg
viewFormInput opts =
    Components.Form.input
        { id = fromFieldToInputType opts.field
        , field = opts.field
        , label = fromFieldToLabel opts.field
        , type_ = fromFieldToInputType opts.field
        , value = opts.value
        , placeholder = fromFieldToLabel opts.field
        , required = True
        , onInput = UserUpdatedInput opts.field
        , helpText = Nothing
        , prefix = Nothing
        }


viewForgotPassword : Html Msg
viewForgotPassword =
    H.div [ A.class "text-right" ]
        [ H.button
            [ A.class "text-sm text-black hover:underline focus:outline-none"
            , A.type_ "button"
            , E.onClick (UserChangedFormVariant ForgotPassword)
            ]
            [ H.text "Forgot password?" ]
        ]


viewSubmitButton : Model -> Html Msg
viewSubmitButton model =
    H.button
        [ A.type_ "submit"
        , A.disabled (isFormDisabled model)
        , A.class
            (if isFormDisabled model then
                "w-full px-4 py-2 bg-gray-300 text-gray-500 rounded-md cursor-not-allowed transition-colors"

             else
                "w-full px-4 py-2 bg-black text-white rounded-md hover:bg-gray-800 focus:outline-none focus:ring-2 focus:ring-black focus:ring-offset-2 transition-colors"
            )
        ]
        [ H.text (fromVariantToLabel model.formVariant) ]


isFormDisabled : Model -> Bool
isFormDisabled model =
    case model.formVariant of
        SignIn ->
            model.isSubmittingForm
                || String.isEmpty model.email
                || String.isEmpty model.password

        SignUp ->
            model.isSubmittingForm
                || String.isEmpty model.email
                || String.isEmpty model.password
                || String.isEmpty model.passwordAgain
                || (model.password /= model.passwordAgain)

        ForgotPassword ->
            model.isSubmittingForm || String.isEmpty model.email

        SetNewPassword _ ->
            model.isSubmittingForm
                || String.isEmpty model.password
                || String.isEmpty model.passwordAgain
                || (model.password /= model.passwordAgain)


fromVariantToLabel : FormVariant -> String
fromVariantToLabel variant =
    case variant of
        SignIn ->
            "Sign In"

        SignUp ->
            "Sign Up"

        ForgotPassword ->
            "Forgot Password"

        SetNewPassword _ ->
            "Set new password"


fromFieldToLabel : Field -> String
fromFieldToLabel field =
    case field of
        Email ->
            "Email address"

        Password ->
            "Password"

        PasswordAgain ->
            "Confirm password"


fromFieldToInputType : Field -> String
fromFieldToInputType field =
    case field of
        Email ->
            "email"

        Password ->
            "password"

        PasswordAgain ->
            "password"
