module Pages.SignIn exposing (Model, Msg, page)

import Data.Credentials exposing (Credentials)
import Effect exposing (Effect)
import Html
import Http
import Page exposing (Page)
import Route exposing (Route)
import Shared
import View exposing (View)


page : Shared.Model -> Route () -> Page Model Msg
page shared route =
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
subscriptions model =
    Sub.none



-- VIEW


view : Model -> View Msg
view model =
    { title = "Pages.SignIn"
    , body = [ Html.text "/sign-in" ]
    }
