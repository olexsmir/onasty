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
                "{\"slug\":\"the.note-slug\"}"
                    |> D.decodeString Data.Note.decodeCreateResponse
                    |> Expect.equal (Ok { slug = "the.note-slug" })
            )
        ]
