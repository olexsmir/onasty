module Pages.Auth exposing (Model, Msg, Variant, page)

import Api
import Api.Auth
import Auth.User
import Data.Credentials exposing (Credentials)
import Effect exposing (Effect)
import Html exposing (Html)
import Html.Attributes as Attr
import Html.Events
import Http
import Layouts
import Page exposing (Page)
import Route exposing (Route)
import Route.Path
import Shared
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
    , error : Maybe Http.Error
    }


init : Shared.Model -> () -> ( Model, Effect Msg )
init shared _ =
    ( { isSubmittingForm = False
      , email = ""
      , password = ""
      , passwordAgain = ""
      , formVariant = SignIn
      , error = Nothing
      }
    , case shared.user of
        Auth.User.SignedIn _ ->
            Effect.pushRoutePath Route.Path.Home_

        Auth.User.NotSignedIn ->
            Effect.none
    )



-- UPDATE


type Msg
    = UserUpdatedInput Field String
    | UserChangedFormVariant Variant
    | UserClickedSubmit
    | ApiSignInResponded (Result Http.Error Credentials)
    | ApiSignUpResponded (Result Http.Error ())


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
            ( { model | isSubmittingForm = False }
            , Effect.signin credentials
            )

        ApiSignInResponded (Err error) ->
            ( { model | isSubmittingForm = False, error = Just error }, Effect.none )

        ApiSignUpResponded (Ok ()) ->
            -- TODO: show banner with that they have to activate account
            ( { model | isSubmittingForm = False }, Effect.none )

        ApiSignUpResponded (Err error) ->
            ( { model | isSubmittingForm = False, error = Just error }, Effect.none )



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions _ =
    Sub.none



-- VIEW


view : Model -> View Msg
view model =
    { title = "Authentication"
    , body =
        [ Html.div [ Attr.class "center" ]
            -- TODO: add oauth buttons
            [ viewError model.error
            , viewChangeVariant model.formVariant
            , viewForm model
            , viewForgotPassword
            ]
        ]
    }


viewChangeVariant : Variant -> Html Msg
viewChangeVariant variant =
    Html.div [ Attr.class "mb1" ]
        [ Html.button
            [ Attr.disabled (variant == SignIn)
            , Html.Events.onClick (UserChangedFormVariant SignIn)
            ]
            [ Html.text "Sign In" ]
        , Html.button
            [ Attr.disabled (variant == SignUp)
            , Html.Events.onClick (UserChangedFormVariant SignUp)
            ]
            [ Html.text "Sign Up" ]
        ]


viewForm : Model -> Html Msg
viewForm model =
    Html.form [ Html.Events.onSubmit UserClickedSubmit ]
        (case model.formVariant of
            SignIn ->
                [ viewFormInput { field = Email, value = model.email }
                , viewFormInput { field = Password, value = model.password }
                , viewSubmitButton model
                ]

            SignUp ->
                [ viewFormInput { field = Email, value = model.email }
                , viewFormInput { field = Password, value = model.password }
                , viewFormInput { field = PasswordAgain, value = model.passwordAgain }
                , viewSubmitButton model
                ]
        )


viewError : Maybe Http.Error -> Html Msg
viewError maybeError =
    case maybeError of
        Just error ->
            Html.div [ Attr.class "box bad" ]
                [ Html.strong [ Attr.class "block titlebar" ] [ Html.text "Error" ]
                , Html.text (Api.errorToFriendlyMessage error)
                ]

        Nothing ->
            Html.text ""


viewFormInput : { field : Field, value : String } -> Html Msg
viewFormInput opts =
    Html.div [ Attr.class "mb1" ]
        [ Html.label [] [ Html.text (fromFieldToLabel opts.field) ]
        , Html.div []
            [ Html.input
                [ Attr.type_ (fromFieldToInputType opts.field)
                , Attr.value opts.value
                , Html.Events.onInput (UserUpdatedInput opts.field)
                ]
                []
            ]
        ]


viewForgotPassword : Html Msg
viewForgotPassword =
    Html.div []
        [ Html.a
            [ Attr.href "/forgot-password"
            , Attr.class "gray"
            ]
            [ Html.text "Forgot password?" ]
        ]


viewSubmitButton : Model -> Html Msg
viewSubmitButton model =
    Html.div [ Attr.class "mb1" ]
        [ Html.button
            [ Attr.disabled (isFormDisabled model) ]
            [ Html.text (fromVariantToLabel model.formVariant) ]
        ]


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
            "Password again"


fromFieldToInputType : Field -> String
fromFieldToInputType field =
    case field of
        Email ->
            "email"

        Password ->
            "password"

        PasswordAgain ->
            "password"
