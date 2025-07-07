module Api exposing (Error(..), Response(..), errorMessage, is404, isNotVerified)

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


errorMessage : Error -> String
errorMessage error =
    case error of
        HttpError err ->
            err.message

        JsonDecodeError err ->
            err.message


is404 : Error -> Bool
is404 error =
    case error of
        HttpError { reason } ->
            reason == Http.BadStatus 404

        _ ->
            False


isNotVerified : Error -> Bool
isNotVerified error =
    case error of
        HttpError { reason, message } ->
            (reason == Http.BadStatus 400) && String.contains "user is not activated" message

        _ ->
            False
