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



-- INIT


type alias Model =
    { email : String
    , password : String
    , isSubmittingForm : Bool
    , error : Maybe Http.Error
    }


init : Shared.Model -> () -> ( Model, Effect Msg )
init shared _ =
    ( { isSubmittingForm = False
      , email = ""
      , password = ""
      , error = Nothing
      }
    , case shared.credentials of
        Just _ ->
            Effect.pushRoutePath Route.Path.Home_

        Nothing ->
            Effect.none
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
        [ Html.div []
            [ Html.div []
                [ Html.div []
                    [ Html.h1 [] [ Html.text "Sign in" ]
                    , viewError model.error
                    , viewForm model
                    ]
                ]
            ]
        ]
    }


viewForm : Model -> Html Msg
viewForm model =
    Html.form [ Html.Events.onSubmit UserClickedSubmit ]
        [ viewFormInput { field = Email, value = model.email }
        , viewFormInput { field = Password, value = model.password }
        , viewFormControls model
        ]


viewError : Maybe Http.Error -> Html Msg
viewError maybeError =
    case maybeError of
        Just error ->
            Html.div [ Attr.style "color" "red" ]
                [ Html.text (Api.errorToFriendlyMessage error) ]

        Nothing ->
            Html.text ""


viewFormInput : { field : Field, value : String } -> Html Msg
viewFormInput opts =
    Html.div []
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
    Html.div []
        [ Html.button
            [ Attr.disabled model.isSubmittingForm ]
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
