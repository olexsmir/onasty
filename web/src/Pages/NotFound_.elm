module Pages.NotFound_ exposing (Model, Msg, page)

import Effect
import Html as H
import Html.Attributes as A
import Layouts
import Page exposing (Page)
import Route exposing (Route)
import Shared
import View exposing (View)


type alias Model =
    {}


type alias Msg =
    ()


page : Shared.Model -> Route () -> Page Model Msg
page _ _ =
    Page.new
        { init = \_ -> ( {}, Effect.none )
        , update = \_ _ -> ( {}, Effect.none )
        , subscriptions = \_ -> Sub.none
        , view = view
        }
        |> Page.withLayout Layouts.Header


view : Model -> View Msg
view _ =
    { title = "404"
    , body = [ H.div [ A.class "py-8 mx-auto w-64" ] [ H.text "Page not found" ] ]
    }
