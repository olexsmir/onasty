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
        , Components.Form.button
            { text = "Resend verification email"
            , onClick = UserClickedResendActivationEmail
            , disabled = not canClick
            , style = Components.Form.SecondaryDisabled canClick
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
        [ Components.Form.button
            { text = "Sign In"
            , onClick = UserChangedFormVariant SignIn
            , style = Components.Form.Primary (variant == SignIn)
            , disabled = variant == SignIn
            }
        , Components.Form.button
            { text = "Sign Up"
            , disabled = variant == SignUp
            , style = Components.Form.Primary (variant == SignUp)
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
                [ viewFormInput { field = Email, value = model.email, error = validateEmail model.email }
                , viewFormInput { field = Password, value = model.password, error = validatePassword model.password }
                , viewForgotPassword
                , viewSubmitButton model
                ]

            SignUp ->
                [ viewFormInput { field = Email, value = model.email, error = validateEmail model.email }
                , viewFormInput { field = Password, value = model.password, error = validatePassword model.password }
                , viewFormInput { field = PasswordAgain, value = model.passwordAgain, error = validatePasswords model.password model.passwordAgain }
                , viewSubmitButton model
                ]

            ForgotPassword ->
                [ viewFormInput { field = Email, value = model.email, error = validateEmail model.email }
                , viewSubmitButton model
                ]

            SetNewPassword _ ->
                [ viewFormInput { field = Password, value = model.password, error = validatePassword model.password }
                , viewFormInput { field = PasswordAgain, value = model.passwordAgain, error = validatePasswords model.password model.passwordAgain }
                , viewSubmitButton model
                ]
        )


viewFormInput : { field : Field, value : String, error : Maybe String } -> Html Msg
viewFormInput opts =
    Components.Form.input
        { style = Components.Form.Simple
        , id = (fromFieldToFieldInfo opts.field).label
        , error = opts.error
        , label = (fromFieldToFieldInfo opts.field).label
        , type_ = (fromFieldToFieldInfo opts.field).type_
        , placeholder = (fromFieldToFieldInfo opts.field).label
        , onInput = UserUpdatedInput opts.field
        , field = opts.field
        , value = opts.value
        , required = True
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
        , style = Components.Form.Primary (isFormDisabled model)
        , disabled = isFormDisabled model
        }


isFormDisabled : Model -> Bool
isFormDisabled model =
    case model.formVariant of
        SignIn ->
            model.isSubmittingForm
                || (validateEmail model.email /= Nothing)
                || (validatePassword model.password /= Nothing)

        SignUp ->
            model.isSubmittingForm
                || (validateEmail model.email /= Nothing)
                || (validatePassword model.password /= Nothing)
                || (validatePasswords model.password model.passwordAgain /= Nothing)

        ForgotPassword ->
            model.isSubmittingForm || (validateEmail model.email /= Nothing)

        SetNewPassword _ ->
            model.isSubmittingForm
                || (validateEmail model.email /= Nothing)
                || (validatePassword model.password /= Nothing)
                || (validatePasswords model.password model.passwordAgain /= Nothing)


validateEmail : String -> Maybe String
validateEmail email =
    if
        not (String.isEmpty email)
            && (not (String.contains "@" email) || not (String.contains "." email))
    then
        Just "Please enter a valid email address."

    else
        Nothing


validatePassword : String -> Maybe String
validatePassword passwd =
    if not (String.isEmpty passwd) && String.length passwd < 8 then
        Just "Password must be at least 8 characters long."

    else
        Nothing


validatePasswords : String -> String -> Maybe String
validatePasswords passowrd1 password2 =
    if not (String.isEmpty passowrd1) && passowrd1 /= password2 then
        Just "Passwords do not match."

    else
        Nothing


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


fromFieldToFieldInfo : Field -> { label : String, type_ : String }
fromFieldToFieldInfo field =
    case field of
        Email ->
            { label = "Email address", type_ = "email" }

        Password ->
            { label = "Password", type_ = "password" }

        PasswordAgain ->
            { label = "Confirm password", type_ = "password" }
