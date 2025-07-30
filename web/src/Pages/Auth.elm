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
        [ H.div [ A.class "min-h-screen flex items-center justify-center bg-gray-50 p-4" ]
            [ H.div [ A.class "w-full max-w-md bg-white rounded-lg border border-gray-200 shadow-sm" ]
                -- TODO: add oauth buttons
                [ viewBanner model
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
        , Components.Form.btn
            { text = "Resend verification email"
            , onClick = UserClickedResendActivationEmail
            , disabled = not canClick
            , style = Components.Form.BorderedGrayedOut canClick
            }
        , Components.Utils.viewIf (not canClick)
            (H.p [ A.class "text-gray-600 text-xs mt-2" ]
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
    H.div [ A.class "flex [&>*]:flex-1 gap-2" ]
        [ Components.Form.btn
            { text = "Sign In"
            , onClick = UserChangedFormVariant SignIn
            , style = Components.Form.Solid (variant == SignIn)
            , disabled = variant == SignIn
            }
        , Components.Form.btn
            { text = "Sign Up"
            , disabled = variant == SignUp
            , style = Components.Form.Solid (variant == SignUp)
            , onClick = UserChangedFormVariant SignUp
            }
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
                , Components.Form.submitButton
                    { text = "Sign In"
                    , class = "w-full"
                    , style = Components.Form.Solid (isFormDisabled model)
                    , disabled = isFormDisabled model
                    }
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
    Components.Form.submitButton
        { class = "w-full"
        , text = fromVariantToLabel model.formVariant
        , style = Components.Form.Solid (isFormDisabled model)
        , disabled = isFormDisabled model
        }


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
