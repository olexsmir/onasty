module UnitTests.Data.Me exposing (..)

import Data.Me
import Expect
import Json.Decode as Json
import Test exposing (Test, describe, test)


suite : Test
suite =
    describe "Data.Me"
        [ test "decode credentials" <|
            \_ ->
                """
                {
                  "email": "admin@onasty.local",
                  "created_at": "2025-06-06T19:44:17.370068Z"
                }
                """
                    |> Json.decodeString Data.Me.decode
                    |> Expect.ok
        ]
