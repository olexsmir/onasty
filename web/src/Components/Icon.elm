module Components.Icon exposing (IconType(..), view)

import Html as H exposing (Html)
import Html.Attributes as A


type IconType
    = NoteIcon
    | NotFound
    | Warning


view : IconType -> String -> Html msg
view t cls =
    let
        getText img =
            H.img [ A.src ("/static/" ++ img ++ ".svg"), A.class cls ] []
    in
    case t of
        NoteIcon ->
            getText "note-icon"

        NotFound ->
            getText "note-not-found"

        Warning ->
            getText "warning"
