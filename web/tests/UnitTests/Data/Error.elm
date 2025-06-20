module UnitTests.Data.Error exposing (suite)

import Data.Error
import Expect
import Json.Decode as Json
import Test exposing (Test, describe, test)


suite : Test
suite =
    describe "Data.Error"
        [ test "decode" <|
            \_ ->
                """
                {
                    "message": "some kind of an error"
                }
                """
                    |> Json.decodeString Data.Error.decode
                    |> Expect.equal (Ok { message = "some kind of an error" })
        ]
