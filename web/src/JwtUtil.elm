module JwtUtil exposing (isExpired)

import Jwt
import Time


{-| Checks if a JWT token is expired or about to expire.
-}
isExpired : Time.Posix -> String -> Bool
isExpired now token =
    let
        expirationThreshold =
            40 * 1000

        timeDiff =
            getTokenExpiration token
                |> (\expiration -> expiration - Time.posixToMillis now)
    in
    timeDiff <= expirationThreshold


getTokenExpiration : String -> Int
getTokenExpiration token =
    Jwt.getTokenExpirationMillis token |> Result.withDefault 0
