module UnitTests.Data.Credentiala exposing (suite)

import Data.Credentials
import Expect
import Json.Decode as Json
import Test exposing (Test, describe, test)


suite : Test
suite =
    describe "Data.Credentials"
        [ test "decode" <|
            \_ ->
                """
                {
                    "access_token": "access.token.value",
                    "refresh_token": "refresh-token-value"
                }
                """
                    |> Json.decodeString Data.Credentials.decode
                    |> Expect.ok
        ]
