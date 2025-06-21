module Pages.Home_ exposing (Model, Msg, page)

import Effect exposing (Effect)
import Html as H
import Html.Attributes as A
import Html.Events as E
import Layouts
import Page exposing (Page)
import Route exposing (Route)
import Shared
import View exposing (View)


page : Shared.Model -> Route () -> Page Model Msg
page shared _ =
    Page.new
        { init = init shared
        , update = update
        , subscriptions = subscriptions
        , view = view shared
        }
        |> Page.withLayout Layouts.Header



-- INIT


type alias Model =
    {}


init : Shared.Model -> () -> ( Model, Effect Msg )
init _ () =
    ( {}, Effect.none )



-- UPDATE


type Msg
    = NoOp


update : Msg -> Model -> ( Model, Effect Msg )
update msg model =
    case msg of
        NoOp ->
            ( model, Effect.none )



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions _ =
    Sub.none



-- VIEW


view : Shared.Model -> Model -> View Msg
view _ _ =
    { title = "Homepage"
    , body =
        [ H.div [ A.class "w-full max-w-6xl mx-auto" ]
            [ H.p [ E.onClick NoOp ] [ H.text "Hello, world!" ] ]
        ]
    }
