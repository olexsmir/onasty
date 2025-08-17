module UnitTests.Data.Me exposing (suite)

import Data.Me
import Expect
import Json.Decode as Json
import Test exposing (Test, describe, test)


suite : Test
suite =
    describe "Data.Me"
        [ test "decode" <|
            \_ ->
                """
                {
                  "email": "admin@onasty.local",
                  "created_at": "2025-06-06T19:44:17.370068Z",
                  "last_login_at": "2025-07-06T17:15:23.380068Z",
                  "notes_created": 42
                }
                """
                    |> Json.decodeString Data.Me.decode
                    |> Expect.ok
        ]
