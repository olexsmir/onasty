module Api.Endpoint exposing (..)

import Http
import Url.Builder exposing (QueryParameter)


type Endpoint
    = Endpoint String


unwrap : Endpoint -> String
unwrap (Endpoint e) =
    e


url : List String -> List QueryParameter -> Endpoint
url paths params =
    Endpoint <|
        Url.Builder.crossOrigin "http://localhost:8000" ("api" :: "v1" :: paths) params


request :
    { body : Http.Body
    , expect : Http.Expect a
    , headers : List Http.Header
    , method : String
    , timeout : Maybe Float
    , url : Endpoint
    , withCredentials : Bool
    }
    -> Http.Request a
request c =
    Http.request
        { body = c.body
        , expect = c.expect
        , headers = c.headers
        , method = c.method
        , timeout = c.timeout
        , url = unwrap c.url
        , withCredentials = c.withCredentials
        }
