module Api.Note exposing (create, fetchMetadata, get)

import Api
import Data.Note as Note exposing (CreateResponse, Metadata, Note)
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


get :
    { onResponse : Result Api.Error Note -> msg
    , slug : String
    , password : Maybe String
    }
    -> Effect msg
get options =
    let
        body : Http.Body
        body =
            case options.password of
                Just password ->
                    E.object [ ( "password", E.string password ) ]
                        |> Http.jsonBody

                Nothing ->
                    Http.emptyBody
    in
    Effect.sendApiRequest
        { endpoint = "/api/v1/note/" ++ options.slug
        , method = "GET"
        , body = body
        , onResponse = options.onResponse
        , decoder = Note.decode
        }


fetchMetadata :
    { onResponse : Result Api.Error Metadata -> msg
    , slug : String
    }
    -> Effect msg
fetchMetadata options =
    Effect.sendApiRequest
        { endpoint = "/api/v1/note/" ++ options.slug ++ "/meta"
        , method = "GET"
        , body = Http.emptyBody
        , onResponse = options.onResponse
        , decoder = Note.decodeMetadata
        }
