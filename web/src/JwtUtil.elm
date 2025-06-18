module JwtUtil exposing (isExpired)

import Jwt
import Time


{-| Checks if a JWT token is expired or about to expire.
-}
isExpired : Time.Posix -> String -> Bool
isExpired now token =
    let
        expirationThreshold : number
        expirationThreshold =
            40 * 1000

        timeDiff : Int
        timeDiff =
            getTokenExpiration token
                |> (\expiration -> expiration - Time.posixToMillis now)
    in
    timeDiff <= expirationThreshold


{-| Extracts the expiration time (in millis) from a JWT token.
Returns 0 if cannot parse token.
-}
getTokenExpiration : String -> Int
getTokenExpiration token =
    Jwt.getTokenExpirationMillis token
        |> Result.withDefault 0
