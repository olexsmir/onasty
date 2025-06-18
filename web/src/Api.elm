module Api exposing (HttpRequestDetails, Response(..))

import Http
import Json.Decode


type Response value
    = Loading
    | Success value
    | Failure Http.Error


type alias HttpRequestDetails msg =
    { endpoint : String
    , method : String
    , body : Http.Body
    , decoder : Json.Decode.Decoder msg
    , onHttpError : Http.Error -> msg
    }
