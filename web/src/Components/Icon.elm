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
        getHtml img =
            H.img [ A.src ("/static/" ++ img ++ ".svg"), A.class cls ] []
    in
    case t of
        NoteIcon ->
            getHtml "note-icon"

        NotFound ->
            getHtml "note-not-found"

        Warning ->
            getHtml "warning"
