module Api.Auth exposing (signin)

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
