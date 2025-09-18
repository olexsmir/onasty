module Pages.Dashboard exposing (Model, Msg, page)

import Api exposing (Response(..))
import Api.Note
import Auth
import Components.Form
import Components.Utils
import Data.Note exposing (Note)
import Effect exposing (Effect)
import Html as H exposing (Html)
import Html.Attributes as A
import Html.Events as E
import Layouts
import Page exposing (Page)
import Route exposing (Route)
import Route.Path
import Shared
import Time exposing (Posix)
import Time.Format
import View exposing (View)


page : Auth.User -> Shared.Model -> Route () -> Page Model Msg
page _ shared _ =
    Page.new
        { init = init
        , update = update
        , subscriptions = subscriptions
        , view = view shared
        }
        |> Page.withLayout (\_ -> Layouts.Header {})


type alias Model =
    { notes : Api.Response (List Note)
    }


init : () -> ( Model, Effect Msg )
init () =
    ( { notes = Api.Loading }
    , Api.Note.getAll { onResponse = ApiNotesResponded }
    )



-- UPDATE


type Msg
    = UserClickedCreateNewNote
    | UserClickedViewNote String
    | UserClickedDeleteNote String
    | ApiNotesResponded (Result Api.Error (List Note))


update : Msg -> Model -> ( Model, Effect Msg )
update msg model =
    case msg of
        UserClickedCreateNewNote ->
            ( model, Effect.pushRoutePath Route.Path.Home_ )

        UserClickedViewNote slug ->
            ( model, Effect.none )

        UserClickedDeleteNote slug ->
            ( model, Effect.none )

        ApiNotesResponded (Ok notes) ->
            ( { model | notes = Api.Success notes }, Effect.none )

        ApiNotesResponded (Err error) ->
            ( { model | notes = Api.Failure error }, Effect.none )


subscriptions : Model -> Sub Msg
subscriptions _ =
    Sub.none



-- VIEW


view : Shared.Model -> Model -> View Msg
view shared model =
    let
        timeFormat =
            Time.Format.toString shared.timeZone
    in
    { title = "Pages.Dashboard"
    , body =
        [ Components.Utils.commonContainer
            [ H.div [ A.class "w-full max-w-6xl mx-auto" ]
                [ H.div [ A.class "bg-white rounded-lg border border-gray-200 shadow-sm" ]
                    [ -- TODO: view error
                      viewHeader
                    , H.div [ A.class "p-6" ] [ viewNotes model.notes timeFormat ]
                    ]
                ]
            ]
        ]
    }


viewCreateNoteButton : Html Msg
viewCreateNoteButton =
    Components.Form.button
        { text = "Create New Note"
        , onClick = UserClickedCreateNewNote
        , style = Components.Form.PrimaryReverse True
        , disabled = False
        }


viewHeader : Html Msg
viewHeader =
    H.div [ A.class "p-6 pb-4 border-b border-gray-200" ]
        [ H.div [ A.class "flex justify-between items-start" ]
            [ H.div []
                [ H.h1 [ A.class "text-2xl font-bold text-gray-900" ] [ H.text "My notes" ]
                , H.p [ A.class "text-gray-600 mt-2" ] [ H.text "Manage and organize all your created notes" ]
                ]
            , H.div [] [ viewCreateNoteButton ]
            ]
        ]


viewNotes : Api.Response (List Note) -> (Posix -> String) -> Html Msg
viewNotes apiResp timeFormat =
    case apiResp of
        Success notes ->
            if List.isEmpty notes then
                viewEmptyNoteList

            else
                H.div [ A.class "space-y-4" ]
                    [ H.div [ A.class "pb-2 border-b border-gray-200" ]
                        [ H.span [ A.class "text-sm text-gray-600" ] [ H.text (String.fromInt (List.length notes) ++ " note(s) ") ] ]
                    , H.div [] (List.map (\n -> viewNoteCard n timeFormat) notes)
                    ]

        Failure err ->
            H.text ("Something went wrong: " ++ Api.errorMessage err)

        Loading ->
            H.text "Loading notes"


viewNoteCard : Note -> (Posix -> String) -> Html Msg
viewNoteCard note timeFormat =
    H.div
        [ A.class
            (if note.readAt /= Nothing then
                "border rounded-lg p-4 border-red-200 bg-red-50"

             else
                "border rounded-lg p-4 border-gray-200 hover:border-gray-300 transition-colors"
            )
        ]
        [ H.div [ A.class "flex items-start justify-between" ]
            [ H.div [ A.class "flex-1 min-w-0" ]
                [ H.p [ A.class "text-gray-700 text-sm mb-3" ] [ H.text (truncateContent note.content) ]
                , H.div [ A.class "flex flex-wrap items-center gap-4 text-xs text-gray-500 mb-2" ]
                    [ H.div [ A.class "flex items-center gap-1" ]
                        [ H.p []
                            [ H.span [ A.class "font-bold" ] [ H.text "Created " ]
                            , H.span [] [ H.text (timeFormat note.createdAt) ]
                            ]
                        , Components.Utils.viewMaybe note.expiresAt
                            (\e ->
                                H.div [ A.class "flex items-center gap-1" ]
                                    [ H.p []
                                        [ H.span [ A.class "font-bold" ] [ H.text "Expires " ]
                                        , H.span [] [ H.text (timeFormat e) ]
                                        ]
                                    ]
                            )
                        , Components.Utils.viewMaybe note.readAt
                            (\r ->
                                H.div [ A.class "flex items-center gap-1" ]
                                    [ H.p []
                                        [ H.span [ A.class "font-bold" ] [ H.text "Read " ]
                                        , H.span [] [ H.text (timeFormat r) ]
                                        ]
                                    ]
                            )
                        ]
                    ]
                , H.div [ A.class "flex flex-wrap gap-2" ]
                    [ Components.Utils.viewIf note.keepBeforeExpiration
                        (H.span [ A.class "inline-flex items-center gap-1 px-2 py-1 bg-orange-100 text-orange-800 text-xs rounded-full" ]
                            [ H.span [] [ H.text "Burn after reading" ] ]
                        )
                    , Components.Utils.viewMaybe note.readAt
                        (\_ ->
                            H.span [ A.class "inline-flex items-center px-2 py-1 bg-blue-100 text-blue-800 text-xs rounded-full" ]
                                [ H.span [] [ H.text "Read" ] ]
                        )
                    ]
                ]
            , H.div [ A.class "flex items-center gap-2 ml-4" ]
                [ H.button
                    [ A.class "p-2 text-gray-400 hover:text-gray-600 bg-gray-50 hover:bg-gray-100 rounded-md transition-colors"
                    , E.onClick (UserClickedViewNote note.slug)
                    , A.title "View note"
                    , A.type_ "button"
                    ]
                    [ H.text "👁️" ]
                , H.button
                    [ A.class "p-2 text-gray-400 text-red-300 hover:text-red-600 bg-red-50 hover:bg-red-100 rounded-md transition-colors disabled:opacity-50"
                    , E.onClick (UserClickedDeleteNote note.slug)
                    , A.title "Delete note"
                    , A.type_ "button"
                    ]
                    [ H.text "🗑️" ]
                ]
            ]
        ]


truncateContent : String -> String
truncateContent content =
    if String.isEmpty content then
        "<DELETED NOTE>"

    else if String.length content <= 150 then
        content

    else
        String.left 150 content ++ "..."


viewEmptyNoteList : Html msg
viewEmptyNoteList =
    H.text "No notes found"
