module Api.Note exposing (create)

import Api
import Data.Note as Note exposing (CreateResponse)
import Effect exposing (Effect)
import Http
import Json.Encode as E


create :
    { onResponse : Result Api.Error CreateResponse -> msg
    , content : String
    , slug : Maybe String
    , password : Maybe String
    }
    -> Effect msg
create options =
    let
        body : E.Value
        body =
            E.object
                [ ( "content", E.string options.content )
                , case options.slug of
                    Just slug ->
                        ( "slug", E.string slug )

                    Nothing ->
                        ( "slug", E.null )
                , case options.password of
                    Just password ->
                        ( "password", E.string password )

                    Nothing ->
                        ( "password", E.null )
                ]
    in
    Effect.sendApiRequest
        { endpoint = "/api/v1/note"
        , method = "POST"
        , body = Http.jsonBody body
        , onResponse = options.onResponse
        , decoder = Note.decodeCreateResponse
        }
