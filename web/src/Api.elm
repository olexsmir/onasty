module Api exposing (Error(..), Response(..))

import Http
import Json.Decode


type Error
    = HttpError
        { message : String
        , reason : Http.Error
        }
    | JsonDecodeError
        { message : String
        , reason : Json.Decode.Error
        }


type Response value
    = Loading
    | Success value
    | Failure Error
