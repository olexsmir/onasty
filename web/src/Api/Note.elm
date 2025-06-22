module Api.Note exposing (create)

import Api
import Data.Note as Note exposing (CreateResponse)
import Effect exposing (Effect)
import Http
import Json.Encode as E


create :
    { onResponse : Result Api.Error CreateResponse -> msg
    , content : String
    , slug : String
    }
    -> Effect msg
create options =
    let
        body : E.Value
        body =
            E.object
                [ ( "content", E.string options.content ) ]
    in
    Effect.sendApiRequest
        { endpoint = "/api/v1/note"
        , method = "POST"
        , body = Http.jsonBody body
        , onResponse = options.onResponse
        , decoder = Note.decodeCreateResponse
        }
