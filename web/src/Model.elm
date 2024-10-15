module Model exposing (Model, Page(..))

import Api
import Browser.Navigation exposing (Key)


type alias Model =
    { apiResponse : Maybe Api.Response
    , curPage : Page
    , navKey : Key
    }


type Page
    = Home
    | NotFound
