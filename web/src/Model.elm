module Model exposing (Model, Page(..))

import Browser.Navigation exposing (Key)
import Viewer exposing (Viewer)


type alias Model =
    { viewer : Viewer
    , curPage : Page
    , navKey : Key
    }


type Page
    = Home
    | NotFound
