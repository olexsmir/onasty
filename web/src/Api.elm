module Api exposing (Error(..), Response(..), getErrorMessage)

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


getErrorMessage : Error -> String
getErrorMessage error =
    case error of
        HttpError err ->
            err.message

        JsonDecodeError err ->
            err.message
