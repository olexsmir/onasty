module Main exposing (main)

import Browser exposing (Document, UrlRequest)
import Browser.Navigation exposing (Key)
import Model exposing (Model, Page(..))
import Msg exposing (Msg(..))
import Pages.HomePage
import Pages.NotFound
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
init _ url key =
    let
        initModel : Model
        initModel =
            { curPage = Home
            , apiResponse = Nothing
            , navKey = key
            }
    in
    ( initModel, Cmd.none )


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
    case model.curPage of
        Home ->
            { title = "Onasty"
            , body = [ Pages.HomePage.view model ]
            }

        NotFound ->
            { title = "404 Not Found"
            , body = [ Pages.NotFound.view ]
            }
