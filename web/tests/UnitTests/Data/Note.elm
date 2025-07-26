module UnitTests.Data.Note exposing (suite)

import Data.Note
import Expect
import Json.Decode as D
import Test exposing (Test, describe, test)


suite : Test
suite =
    describe "Data.Note"
        [ test "decodeCreateResponse"
            (\_ ->
                """ {"slug":"the.note-slug"} """
                    |> D.decodeString Data.Note.decodeCreateResponse
                    |> Expect.ok
            )
        , test "decodeMetadata"
            (\_ ->
                """
                {
                    "created_at": "2023-10-01T12:00:00Z",
                    "has_password": false
                }
                """
                    |> D.decodeString Data.Note.decodeMetadata
                    |> Expect.ok
            )
        ]
