module Pages.SignIn exposing (Model, Msg, page)

import Api
import Api.Auth
import Data.Credentials exposing (Credentials)
import Effect exposing (Effect)
import Html exposing (Html)
import Html.Attributes as Attr
import Html.Events
import Http
import Page exposing (Page)
import Route exposing (Route)
import Shared
import View exposing (View)


page : Shared.Model -> Route () -> Page Model Msg
page _ _ =
    Page.new
        { init = init
        , update = update
        , subscriptions = subscriptions
        , view = view
        }



-- INIT


type alias Model =
    { email : String
    , password : String
    , isSubmittingForm : Bool
    , error : Maybe Http.Error
    }


init : () -> ( Model, Effect Msg )
init () =
    ( { isSubmittingForm = False
      , email = ""
      , password = ""
      , error = Nothing
      }
    , Effect.none
    )



-- UPDATE


type Msg
    = UserUpdatedInput Field String
    | UserClickedSubmit
    | ApiSignInResponded (Result Http.Error Credentials)


type Field
    = Email
    | Password


update : Msg -> Model -> ( Model, Effect Msg )
update msg model =
    case msg of
        UserClickedSubmit ->
            ( { model | isSubmittingForm = True }
            , Api.Auth.signin
                { onResponse = ApiSignInResponded
                , email = model.email
                , password = model.password
                }
            )

        UserUpdatedInput Email email ->
            ( { model | email = email }, Effect.none )

        UserUpdatedInput Password password ->
            ( { model | password = password }, Effect.none )

        ApiSignInResponded (Ok credentials) ->
            ( { model | isSubmittingForm = False }
            , Effect.signin credentials
            )

        ApiSignInResponded (Err error) ->
            ( { model | isSubmittingForm = False, error = Just error }
            , Effect.none
            )



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions _ =
    Sub.none



-- VIEW


view : Model -> View Msg
view model =
    { title = "Sign-in"
    , body =
        [ Html.div [ Attr.class "columns is-mobile is-centered" ]
            [ Html.div [ Attr.class "column is-narrow" ]
                [ Html.div [ Attr.class "section" ]
                    [ Html.h1 [ Attr.class "title" ] [ Html.text "Sign in" ]
                    , viewError model.error
                    , viewForm model
                    ]
                ]
            ]
        ]
    }


viewForm : Model -> Html Msg
viewForm model =
    Html.form [ Attr.class "box", Html.Events.onSubmit UserClickedSubmit ]
        [ viewFormInput { field = Email, value = model.email }
        , viewFormInput { field = Password, value = model.password }
        , viewFormControls model
        ]


viewError : Maybe Http.Error -> Html Msg
viewError maybeError =
    case maybeError of
        Just error ->
            Html.div [ Attr.class "is-danger" ]
                [ Html.text (Api.errorToFriendlyMessage error) ]

        Nothing ->
            Html.text ""


viewFormInput : { field : Field, value : String } -> Html Msg
viewFormInput opts =
    Html.div [ Attr.class "field" ]
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


viewFormControls : Model -> Html Msg
viewFormControls model =
    Html.div [ Attr.class "control" ]
        [ Html.button
            [ Attr.class "button is-link"
            , Attr.disabled model.isSubmittingForm
            , Attr.classList [ ( "is-loading", model.isSubmittingForm ) ]
            ]
            [ Html.text "Sign In" ]
        ]


fromFieldToLabel : Field -> String
fromFieldToLabel field =
    case field of
        Email ->
            "Email address"

        Password ->
            "Password"


fromFieldToInputType : Field -> String
fromFieldToInputType field =
    case field of
        Email ->
            "email"

        Password ->
            "password"
