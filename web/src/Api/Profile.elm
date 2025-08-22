module Api.Profile exposing (changePassword, me, requestEmailChange)

import Api
import Data.Me as Me exposing (Me)
import Effect exposing (Effect)
import Http
import Json.Decode as Decode
import Json.Encode as E


me : { onResponse : Result Api.Error Me -> msg } -> Effect msg
me options =
    Effect.sendApiRequest
        { endpoint = "/api/v1/me"
        , method = "GET"
        , body = Http.emptyBody
        , onResponse = options.onResponse
        , decoder = Me.decode
        }


requestEmailChange : { onResponse : Result Api.Error () -> msg, newEmail : String } -> Effect msg
requestEmailChange { onResponse, newEmail } =
    Effect.sendApiRequest
        { endpoint = "/api/v1/auth/change-email"
        , method = "POST"
        , body = E.object [ ( "new_email", E.string newEmail ) ] |> Http.jsonBody
        , onResponse = onResponse
        , decoder = Decode.succeed ()
        }


changePassword : { onResponse : Result Api.Error () -> msg, currentPassword : String, newPassword : String } -> Effect msg
changePassword { onResponse, currentPassword, newPassword } =
    Effect.sendApiRequest
        { endpoint = "/api/v1/auth/change-password"
        , method = "POST"
        , body =
            Http.jsonBody <|
                E.object
                    [ ( "current_password", E.string currentPassword )
                    , ( "new_password", E.string newPassword )
                    ]
        , onResponse = onResponse
        , decoder = Decode.succeed ()
        }
