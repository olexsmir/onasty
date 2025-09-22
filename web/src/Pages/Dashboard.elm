module Pages.Dashboard exposing (Model, Msg, page)

import Api exposing (Response(..))
import Api.Note
import Auth
import Components.Box
import Components.Form
import Components.Utils
import Data.Note exposing (Note)
import Effect exposing (Effect)
import Html as H exposing (Html)
import Html.Attributes as A
import Html.Events as E
import Layouts
import Page exposing (Page)
import Ports
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
    , noteToDeleteSlug : Maybe String
    , apiError : Maybe Api.Error
    }


init : () -> ( Model, Effect Msg )
init () =
    ( { notes = Api.Loading
      , noteToDeleteSlug = Nothing
      , apiError = Nothing
      }
    , Api.Note.getAll { onResponse = ApiNotesResponded }
    )



-- UPDATE


type Msg
    = UserClickedCreateNewNote
    | UserClickedViewNote String
    | UserClickedDeleteNote String
    | UserConfirmedDeleteion Bool
    | ApiNotesResponded (Result Api.Error (List Note))
    | ApiNoteDeleted (Result Api.Error ())


update : Msg -> Model -> ( Model, Effect Msg )
update msg model =
    case msg of
        UserClickedCreateNewNote ->
            ( model, Effect.pushRoutePath Route.Path.Home_ )

        UserClickedViewNote slug ->
            ( model, Effect.pushRoutePath (Route.Path.Secret_Slug_ { slug = slug }) )

        UserClickedDeleteNote slug ->
            ( { model | noteToDeleteSlug = Just slug }
            , Effect.confirmRequest "Are you sure you want to delete this note?"
            )

        UserConfirmedDeleteion ok ->
            case ( ok, model.noteToDeleteSlug ) of
                ( True, Just slug ) ->
                    let
                        newNotes =
                            case model.notes of
                                Success notes ->
                                    Success (List.filter (\n -> n.slug /= slug) notes)

                                _ ->
                                    model.notes
                    in
                    ( { model | notes = newNotes, noteToDeleteSlug = Nothing }
                    , Api.Note.delete { onResponse = ApiNoteDeleted, slug = slug }
                    )

                _ ->
                    ( { model | noteToDeleteSlug = Nothing }, Effect.none )

        ApiNotesResponded (Ok notes) ->
            ( { model | notes = Api.Success notes }, Effect.none )

        ApiNotesResponded (Err error) ->
            ( { model | notes = Api.Failure error }, Effect.none )

        ApiNoteDeleted (Ok _) ->
            ( { model | apiError = Nothing }, Effect.none )

        ApiNoteDeleted (Err err) ->
            ( { model | apiError = Just err }, Effect.none )


subscriptions : Model -> Sub Msg
subscriptions _ =
    Ports.confirmResponse UserConfirmedDeleteion



-- VIEW


view : Shared.Model -> Model -> View Msg
view shared model =
    let
        timeFormat =
            Time.Format.toString shared.timeZone
    in
    { title = "Dashboard"
    , body =
        [ Components.Utils.commonContainer
            [ H.div [ A.class "w-full max-w-6xl mx-auto" ]
                [ H.div [ A.class "bg-white rounded-lg border border-gray-200 shadow-sm" ]
                    [ Components.Utils.viewMaybe model.apiError (\e -> Components.Box.error (Api.errorMessage e))
                    , viewHeader
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
    let
        viewNoteTime text maybeTime =
            Components.Utils.viewMaybe maybeTime
                (\r ->
                    H.div [ A.class "flex items-center" ]
                        [ H.p []
                            [ H.span [ A.class "font-bold" ] [ H.text text ]
                            , H.span [] [ H.text (timeFormat r) ]
                            ]
                        ]
                )

        viewNoteBadges text cond colorClasses =
            Components.Utils.viewIf cond
                (H.span
                    [ A.class ("inline-flex items-center gap-1 px-2 py-1 text-xs rounded-full " ++ colorClasses) ]
                    [ H.span [] [ H.text text ] ]
                )
    in
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
                    [ H.div [ A.class "items-center" ]
                        [ H.p []
                            [ H.span [ A.class "font-bold" ] [ H.text "Created " ]
                            , H.span [] [ H.text (timeFormat note.createdAt) ]
                            ]
                        , viewNoteTime "Read " note.readAt
                        , viewNoteTime "Expires " note.expiresAt
                        ]
                    ]
                , H.div [ A.class "flex flex-wrap gap-2" ]
                    [ viewNoteBadges "Burn after reading" note.keepBeforeExpiration "bg-orange-100 text-orange-800"
                    , viewNoteBadges "Has password" note.hasPassword "bg-blue-100 text-blue-800"
                    , viewNoteBadges "Read" (note.readAt /= Nothing) "bg-red-100 text-red-100"
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
