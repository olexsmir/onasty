module Api.Note exposing (create)

import Api
import Data.Note as Note exposing (CreateResponse)
import Effect exposing (Effect)
import Http
import ISO8601
import Json.Encode as E
import Time exposing (Posix)


create :
    { onResponse : Result Api.Error CreateResponse -> msg
    , content : String
    , slug : Maybe String
    , password : Maybe String
    , expiresAt : Posix
    , burnBeforeExpiration : Bool
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
                , ( "burn_before_expiration", E.bool options.burnBeforeExpiration )
                , if options.expiresAt == Time.millisToPosix 0 then
                    ( "expires_at", E.null )

                  else
                    ( "expires_at"
                    , options.expiresAt
                        |> ISO8601.fromPosix
                        |> ISO8601.toString
                        |> E.string
                    )
                ]
    in
    Effect.sendApiRequest
        { endpoint = "/api/v1/note"
        , method = "POST"
        , body = Http.jsonBody body
        , onResponse = options.onResponse
        , decoder = Note.decodeCreateResponse
        }
