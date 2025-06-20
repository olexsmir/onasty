module Api.Auth exposing (refreshToken, resendVerificationEmail, signin, signup)

import Api
import Data.Credentials as Credentials exposing (Credentials)
import Effect exposing (Effect)
import Http
import Json.Decode as Decode
import Json.Encode as Encode


signin :
    { onResponse : Result Api.Error Credentials -> msg
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


signup :
    { onResponse : Result Api.Error () -> msg
    , email : String
    , password : String
    }
    -> Effect msg
signup options =
    let
        body : Encode.Value
        body =
            Encode.object
                [ ( "email", Encode.string options.email )
                , ( "password", Encode.string options.password )
                ]
    in
    Effect.sendApiRequest
        { endpoint = "/api/v1/auth/signup"
        , method = "POST"
        , body = Http.jsonBody body
        , onResponse = options.onResponse
        , decoder = Decode.succeed ()
        }


refreshToken :
    { onResponse : Result Api.Error Credentials -> msg
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


resendVerificationEmail :
    { onResponse : Result Api.Error () -> msg
    , email : String
    , password : String
    }
    -> Effect msg
resendVerificationEmail options =
    let
        body : Encode.Value
        body =
            Encode.object
                [ ( "email", Encode.string options.email )
                , ( "password", Encode.string options.password )
                ]
    in
    Effect.sendApiRequest
        { endpoint = "/api/v1/auth/resend-verification-email"
        , method = "POST"
        , body = Http.jsonBody body
        , onResponse = options.onResponse
        , decoder = Decode.succeed ()
        }
