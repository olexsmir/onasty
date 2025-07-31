module Api.Auth exposing (forgotPassword, refreshToken, resendVerificationEmail, resetPassword, signin, signup)

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


refreshToken : { onResponse : Result Api.Error Credentials -> msg, refreshToken : String } -> Effect msg
refreshToken options =
    Effect.sendApiRequest
        { endpoint = "/api/v1/auth/refresh-tokens"
        , method = "POST"
        , body = Encode.object [ ( "refresh_token", Encode.string options.refreshToken ) ] |> Http.jsonBody
        , onResponse = options.onResponse
        , decoder = Credentials.decode
        }


forgotPassword : { onResponse : Result Api.Error () -> msg, email : String } -> Effect msg
forgotPassword options =
    Effect.sendApiRequest
        { endpoint = "/api/v1/auth/reset-password"
        , method = "POST"
        , body = Encode.object [ ( "email", Encode.string options.email ) ] |> Http.jsonBody
        , onResponse = options.onResponse
        , decoder = Decode.succeed ()
        }


resetPassword : { onResponse : Result Api.Error () -> msg, token : String, password : String } -> Effect msg
resetPassword options =
    Effect.sendApiRequest
        { endpoint = "/api/v1/auth/reset-password/" ++ options.token
        , method = "POST"
        , body = Encode.object [ ( "password", Encode.string options.password ) ] |> Http.jsonBody
        , onResponse = options.onResponse
        , decoder = Decode.succeed ()
        }


resendVerificationEmail : { onResponse : Result Api.Error () -> msg, email : String } -> Effect msg
resendVerificationEmail options =
    Effect.sendApiRequest
        { endpoint = "/api/v1/auth/resend-verification-email"
        , method = "POST"
        , body = Encode.object [ ( "email", Encode.string options.email ) ] |> Http.jsonBody
        , onResponse = options.onResponse
        , decoder = Decode.succeed ()
        }
