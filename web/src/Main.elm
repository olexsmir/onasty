module Main exposing (main)

import Browser exposing (Document, UrlRequest)
import Browser.Navigation exposing (Key)
import Html as H
import Html.Attributes as A
import Model exposing (Model)
import Msg exposing (Msg(..))
import Url exposing (Url)


type alias Flags =
    {}


main : Program Flags Model Msg
main =
    Browser.application
        { init = init
        , update = update
        , onUrlChange = onUrlChange
        , onUrlRequest = onUrlRequest
        , view = view
        , subscriptions = \_ -> Sub.none
        }


{-| the functions that called when elm first runs
-}
init : Flags -> Url -> Key -> ( Model, Cmd Msg )
init _ url _ =
    let
        _ =
            Debug.log "url" url
    in
    ( Model "User", Cmd.none )


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        NoOp ->
            ( model, Cmd.none )


onUrlRequest : UrlRequest -> Msg
onUrlRequest url =
    NoOp


onUrlChange : Url -> Msg
onUrlChange url =
    NoOp


view : Model -> Document Msg
view model =
    { title = "Onasty"
    , body =
        [ H.div []
            [ H.text "Hello" ]
        ]
    }
