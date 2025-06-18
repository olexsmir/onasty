module Api exposing (HttpRequestDetails, Response(..), errorToFriendlyMessage)

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


errorToFriendlyMessage : Http.Error -> String
errorToFriendlyMessage httpError =
    case httpError of
        Http.BadUrl _ ->
            "This page requested a bad URL"

        Http.Timeout ->
            "Request took too long to respond"

        Http.NetworkError ->
            "Could not connect to the API"

        Http.BadStatus code ->
            case code of
                404 ->
                    "Not found"

                401 ->
                    "Unauthorized"

                _ ->
                    "API returned an error code"

        Http.BadBody _ ->
            "Unexpected response from API"
