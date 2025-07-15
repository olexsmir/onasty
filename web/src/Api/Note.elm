module Api.Note exposing (create, get, getMetadata)

import Api
import Data.Note as Note exposing (CreateResponse, Metadata, Note)
import Effect exposing (Effect)
import Http
import Iso8601
import Json.Encode as E
import Time exposing (Posix)
import Url


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
        encodeMaybe : Maybe a -> String -> (a -> E.Value) -> ( String, E.Value )
        encodeMaybe maybe field value =
            case maybe of
                Just data ->
                    ( field, value data )

                Nothing ->
                    ( field, E.null )

        body =
            E.object
                [ ( "content", E.string options.content )
                , encodeMaybe options.slug "slug" E.string
                , encodeMaybe options.password "password" E.string
                , ( "burn_before_expiration", E.bool options.burnBeforeExpiration )
                , if options.expiresAt == Time.millisToPosix 0 then
                    ( "expires_at", E.null )

                  else
                    ( "expires_at", options.expiresAt |> Iso8601.fromTime |> E.string )
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
    case options.password of
        Just passwd ->
            Effect.sendApiRequest
                { endpoint = "/api/v1/note/" ++ options.slug ++ "/view"
                , method = "POST"
                , body = E.object [ ( "password", E.string passwd ) ] |> Http.jsonBody
                , onResponse = options.onResponse
                , decoder = Note.decode
                }

        Nothing ->
            Effect.sendApiRequest
                { endpoint = "/api/v1/note/" ++ options.slug
                , method = "GET"
                , body = Http.emptyBody
                , onResponse = options.onResponse
                , decoder = Note.decode
                }


getMetadata :
    { onResponse : Result Api.Error Metadata -> msg
    , slug : String
    }
    -> Effect msg
getMetadata options =
    Effect.sendApiRequest
        { endpoint = "/api/v1/note/" ++ options.slug ++ "/meta"
        , method = "GET"
        , body = Http.emptyBody
        , onResponse = options.onResponse
        , decoder = Note.decodeMetadata
        }
