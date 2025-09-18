module Pages.Oauth.Callback exposing (Model, Msg, page)

import Components.Box
import Components.Utils
import Dict exposing (Dict)
import Effect exposing (Effect)
import Layouts
import Page exposing (Page)
import Route exposing (Route)
import Shared
import View exposing (View)


type alias Msg =
    {}


page : Shared.Model -> Route () -> Page Model Msg
page _ route =
    Page.new
        { init = init route.query
        , update = \_ m -> ( m, Effect.none )
        , subscriptions = \_ -> Sub.none
        , view = view
        }
        |> Page.withLayout (\_ -> Layouts.Header {})


type alias Model =
    { error : String }


init : Dict String String -> () -> ( Model, Effect Msg )
init query () =
    case
        ( Dict.get "access_token" query
        , Dict.get "refresh_token" query
        , Dict.get "error" query
        )
    of
        ( Just at, Just rt, _ ) ->
            ( { error = "" }, Effect.signin { accessToken = at, refreshToken = rt } )

        ( _, _, Just error ) ->
            ( { error = error }, Effect.none )

        _ ->
            ( { error = "Invalid server response" }, Effect.none )


view : Model -> View msg
view model =
    { title = "Oauth"
    , body =
        [ Components.Utils.commonContainer
            [ Components.Box.error model.error ]
        ]
    }
