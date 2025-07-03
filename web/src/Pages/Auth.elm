module Pages.Auth exposing (Model, Msg, Variant, page)

import Api
import Api.Auth
import Auth.User
import Components.Error
import Components.Form
import Data.Credentials exposing (Credentials)
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
page shared _ =
    Page.new
        { init = init shared
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
    , formVariant : Variant
    , gotSignedUp : Bool
    , lastClicked : Maybe Posix
    , apiError : Maybe Api.Error
    , now : Maybe Posix
    }


init : Shared.Model -> () -> ( Model, Effect Msg )
init shared _ =
    ( { email = ""
      , password = ""
      , passwordAgain = ""
      , isSubmittingForm = False
      , formVariant = SignIn
      , gotSignedUp = False
      , lastClicked = Nothing
      , apiError = Nothing
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
    | UserChangedFormVariant Variant
    | UserClickedSubmit
    | UserClickedResendActivationEmail
    | ApiSignInResponded (Result Api.Error Credentials)
    | ApiSignUpResponded (Result Api.Error ())
    | ApiResendVerificationEmail (Result Api.Error ())


type Field
    = Email
    | Password
    | PasswordAgain


type Variant
    = SignIn
    | SignUp


update : Msg -> Model -> ( Model, Effect Msg )
update msg model =
    case msg of
        Tick now ->
            ( { model | now = Just now }, Effect.none )

        UserClickedSubmit ->
            ( { model | isSubmittingForm = True, apiError = Nothing }
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
            )

        UserClickedResendActivationEmail ->
            ( { model | lastClicked = model.now }
            , Api.Auth.resendVerificationEmail
                { onResponse = ApiResendVerificationEmail
                , email = model.email
                , password = model.password
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

        ApiSignInResponded (Err error) ->
            ( { model | isSubmittingForm = False, apiError = Just error }, Effect.none )

        ApiSignUpResponded (Ok ()) ->
            ( { model | isSubmittingForm = False, gotSignedUp = True }, Effect.none )

        ApiSignUpResponded (Err error) ->
            ( { model | isSubmittingForm = False, apiError = Just error }, Effect.none )

        ApiResendVerificationEmail (Ok ()) ->
            ( { model | apiError = Nothing }, Effect.none )

        ApiResendVerificationEmail (Err err) ->
            ( { model | apiError = Just err }, Effect.none )



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions model =
    if model.gotSignedUp then
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
                , viewHeader model.formVariant
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
    case ( model.apiError, model.gotSignedUp ) of
        ( Just error, False ) ->
            Components.Error.error (Api.errorMessage error)

        ( Nothing, True ) ->
            viewBannerSuccess model.now model.lastClicked

        _ ->
            H.text ""


viewBannerSuccess : Maybe Posix -> Maybe Posix -> Html Msg
viewBannerSuccess now lastClicked =
    let
        buttonClassesBase =
            "w-full px-4 py-2 rounded-md focus:outline-none focus:ring-2 focus:ring-black focus:ring-offset-2 transition-colors mt-3"

        buttonClasses active =
            if active then
                buttonClassesBase ++ " border border-gray-300 text-gray-700 hover:bg-gray-50"

            else
                buttonClassesBase ++ " border border-gray-300 text-gray-400 cursor-not-allowed"

        timeLeftSeconds : Int
        timeLeftSeconds =
            case ( now, lastClicked ) of
                ( Just now_, Just last ) ->
                    let
                        remainingMs =
                            30 * 1000 - (Time.posixToMillis now_ - Time.posixToMillis last)
                    in
                    if remainingMs > 0 then
                        remainingMs // 1000

                    else
                        0

                _ ->
                    0

        canClick : Bool
        canClick =
            timeLeftSeconds == 0
    in
    H.div [ A.class "bg-green-50 border border-green-200 rounded-md p-4 mb-4" ]
        [ H.div [ A.class "font-medium text-green-800 mb-2" ] [ H.text "Check your email!" ]
        , H.p [ A.class "text-green-800 text-sm" ] [ H.text "We've sent you a verification link. Please check your email and click the link to activate your account." ]
        , H.button
            [ A.class (buttonClasses canClick)
            , E.onClick UserClickedResendActivationEmail
            , A.disabled (not canClick)
            ]
            [ H.text "Resend verification email" ]
        , if canClick then
            H.text ""

          else
            H.p [ A.class "text-gray-600 text-xs mt-2" ]
                [ H.text
                    ("You can request a new verification email in "
                        ++ String.fromInt timeLeftSeconds
                        ++ " seconds."
                    )
                ]
        ]


viewHeader : Variant -> Html Msg
viewHeader variant =
    let
        ( title, description ) =
            case variant of
                SignIn ->
                    ( "Welcome Back", "Enter your credentials to access your account" )

                SignUp ->
                    ( "Create Account", "Enter your information to create your account" )
    in
    H.div [ A.class "p-6 pb-4" ]
        [ H.h1 [ A.class "text-2xl font-bold text-center mb-2" ] [ H.text title ]
        , H.p [ A.class "text-center text-gray-600 text-sm" ] [ H.text description ]
        ]


viewChangeVariant : Variant -> Html Msg
viewChangeVariant variant =
    let
        base =
            "flex-1 px-4 py-2 rounded-md font-medium transition-colors"

        buttonClasses : Bool -> String
        buttonClasses active =
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
        )


viewFormInput : { field : Field, value : String } -> Html Msg
viewFormInput opts =
    Components.Form.input
        { id = fromFieldToInputType opts.field
        , field = opts.field
        , label = fromFieldToLabel opts.field
        , type_ = fromFieldToInputType opts.field
        , value = opts.value
        , prefix = Nothing
        , placeholder = fromFieldToLabel opts.field
        , required = True
        , onInput = UserUpdatedInput opts.field
        , helpText = Nothing
        }


viewForgotPassword : Html Msg
viewForgotPassword =
    H.div [ A.class "text-right" ]
        [ H.button
            [ A.class "text-sm text-black hover:underline focus:outline-none"
            , A.type_ "button"

            -- TODO: implement forgot password
            -- , E.onClick (UserChangedFormVariant ForgotPassword)
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


fromVariantToLabel : Variant -> String
fromVariantToLabel variant =
    case variant of
        SignIn ->
            "Sign In"

        SignUp ->
            "Sign Up"


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
