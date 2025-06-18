module Api.Auth exposing (refreshToken, signin)

import Data.Credentials as Credentials exposing (Credentials)
import Effect exposing (Effect)
import Http
import Json.Encode as Encode


signin :
    { onResponse : Result Http.Error Credentials -> msg
    , email : String
    , password : String
    }
    -> Effect msg
signin options =
    let
        body : Encode.Value
        body =
            Encode.object
                [ ( "email", Encode.string options.email )
                , ( "password", Encode.string options.password )
                ]
    in
    Effect.sendApiRequest
        { endpoint = "/api/v1/auth/signin"
        , method = "POST"
        , body = Http.jsonBody body
        , onResponse = options.onResponse
        , decoder = Credentials.decode
        }


refreshToken :
    { onResponse : Result Http.Error Credentials -> msg
    , refreshToken : String
    }
    -> Effect msg
refreshToken options =
    let
        body : Encode.Value
        body =
            Encode.object
                [ ( "refresh_token", Encode.string options.refreshToken ) ]
    in
    Effect.sendApiRequest
        { endpoint = "/api/v1/auth/refresh-tokens"
        , method = "POST"
        , body = Http.jsonBody body
        , onResponse = options.onResponse
        , decoder = Credentials.decode
        }
