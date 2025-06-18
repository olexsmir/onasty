module Api.Me exposing (get)

import Data.Me as Me exposing (Me)
import Effect exposing (Effect)
import Http


get : { onResponse : Result Http.Error Me -> msg } -> Effect msg
get options =
    Effect.sendApiRequest
        { endpoint = "/api/v1/me"
        , method = "GET"
        , body = Http.emptyBody
        , onResponse = options.onResponse
        , decoder = Me.decode
        }
