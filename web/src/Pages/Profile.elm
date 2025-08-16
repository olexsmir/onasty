module Pages.Profile exposing (Model, Msg, page)

import Api
import Api.Me
import Auth
import Data.Me exposing (Me)
import Effect exposing (Effect)
import Html exposing (Html)
import Layouts
import Page exposing (Page)
import Route exposing (Route)
import Shared
import Time.Format as T
import View exposing (View)


page : Auth.User -> Shared.Model -> Route () -> Page Model Msg
page _ shared _ =
    Page.new
        { init = init shared
        , update = update
        , subscriptions = subscriptions
        , view = view shared
        }
        |> Page.withLayout (\_ -> Layouts.Header {})



-- INIT


type alias Model =
    { me : Api.Response Me }


init : Shared.Model -> () -> ( Model, Effect Msg )
init _ () =
    ( { me = Api.Loading }
    , Api.Me.get { onResponse = ApiMeResponded }
    )



-- UPDATE


type Msg
    = ApiMeResponded (Result Api.Error Me)


update : Msg -> Model -> ( Model, Effect Msg )
update msg model =
    case msg of
        ApiMeResponded (Ok userData) ->
            ( { model | me = Api.Success userData }, Effect.none )

        ApiMeResponded (Err error) ->
            ( { model | me = Api.Failure error }, Effect.none )



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions _ =
    Sub.none



-- VIEW


view : Shared.Model -> Model -> View Msg
view shared model =
    { title = "Profile"
    , body = [ viewProfileContent shared model.me ]
    }


viewProfileContent : Shared.Model -> Api.Response Me -> Html Msg
viewProfileContent shared userResponse =
    case userResponse of
        Api.Loading ->
            Html.text "Loading..."

        Api.Success user ->
            viewUserDetails shared user

        Api.Failure err ->
            Html.text (Api.errorMessage err)


viewUserDetails : Shared.Model -> Me -> Html Msg
viewUserDetails shared me =
    Html.div []
        [ Html.p [] [ Html.text ("Email: " ++ me.email) ]
        , Html.p [] [ Html.text ("Joined: " ++ T.toString shared.timeZone me.createdAt) ]
        ]
