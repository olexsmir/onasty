module Pages.Secret.Slug_ exposing (Model, Msg, PageVariant, page)

import Api
import Api.Note
import Components.Box
import Components.Icon as Icon
import Components.Utils
import Data.Note exposing (Metadata, Note)
import Effect exposing (Effect)
import Html as H exposing (Html)
import Html.Attributes as A
import Html.Events as E
import Layouts
import Page exposing (Page)
import Route exposing (Route)
import Shared
import Time exposing (Zone)
import Time.Format as T
import View exposing (View)


page : Shared.Model -> Route { slug : String } -> Page Model Msg
page shared route =
    Page.new
        { init = init route.params.slug
        , update = update
        , subscriptions = subscriptions
        , view = view shared
        }
        |> Page.withLayout (\_ -> Layouts.Header {})



-- INIT


type PageVariant
    = RequestNote
    | ShowNote (Api.Response Note)
    | NotFound


type alias Model =
    { page : PageVariant
    , metadata : Api.Response Metadata
    , slug : String
    , password : Maybe String
    }


init : String -> () -> ( Model, Effect Msg )
init slug () =
    ( { page = RequestNote
      , metadata = Api.Loading
      , slug = slug
      , password = Nothing
      }
    , Api.Note.getMetadata
        { onResponse = ApiGetMetadataResponded
        , slug = slug
        }
    )



-- UPDATE


type Msg
    = UserClickedViewNote
    | UserClickedCopyContent
    | UserUpdatedPassword String
    | ApiGetNoteResponded (Result Api.Error Note)
    | ApiGetMetadataResponded (Result Api.Error Metadata)


update : Msg -> Model -> ( Model, Effect Msg )
update msg model =
    case msg of
        UserClickedViewNote ->
            ( { model | page = ShowNote Api.Loading }
            , Api.Note.get
                { onResponse = ApiGetNoteResponded
                , password = model.password
                , slug = model.slug
                }
            )

        UserClickedCopyContent ->
            case model.page of
                ShowNote (Api.Success note) ->
                    ( model, Effect.sendToClipboard note.content )

                _ ->
                    ( model, Effect.none )

        UserUpdatedPassword password ->
            ( { model | password = Just password }, Effect.none )

        ApiGetNoteResponded (Ok note) ->
            ( { model | page = ShowNote (Api.Success note) }, Effect.none )

        ApiGetNoteResponded (Err error) ->
            ( { model | page = ShowNote (Api.Failure error) }, Effect.none )

        ApiGetMetadataResponded (Ok metadata) ->
            ( { model | metadata = Api.Success metadata }, Effect.none )

        ApiGetMetadataResponded (Err error) ->
            ( { model | page = NotFound, metadata = Api.Failure error }, Effect.none )


subscriptions : Model -> Sub Msg
subscriptions _ =
    Sub.none



-- VIEW


view : Shared.Model -> Model -> View Msg
view shared model =
    { title = "View note"
    , body =
        [ Components.Utils.commonContainer
            (case model.metadata of
                Api.Success metadata ->
                    viewPage shared.timeZone model.slug model.page metadata model.password

                Api.Loading ->
                    [ viewHeader { title = "View note", subtitle = "Loading note metadata..." }
                    , viewOpenNote { slug = model.slug, hasPassword = False, password = Nothing, isLoading = True }
                    ]

                Api.Failure error ->
                    [ viewHeader { title = "Note Not Found", subtitle = "The note you're looking for doesn't exist or has expired" }
                    , if Api.is404 error then
                        viewNoteNotFound

                      else
                        Components.Box.error (Api.errorMessage error)
                    ]
            )
        ]
    }


viewPage : Zone -> String -> PageVariant -> Metadata -> Maybe String -> List (Html Msg)
viewPage zone slug variant metadata password =
    case variant of
        RequestNote ->
            [ viewHeader { title = "View note", subtitle = "Click the button below to view the note content" }
            , viewOpenNote { slug = slug, hasPassword = metadata.hasPassword, password = password, isLoading = False }
            ]

        ShowNote apiResp ->
            case apiResp of
                Api.Success note ->
                    [ viewShowNoteHeader zone slug note
                    , viewNoteContent note
                    ]

                Api.Loading ->
                    [ viewHeader { title = "View note", subtitle = "Click the button below to view the note content" }
                    , viewOpenNote { slug = slug, hasPassword = metadata.hasPassword, password = password, isLoading = True }
                    ]

                Api.Failure _ ->
                    [ viewHeader { title = "Note Not Found", subtitle = "The note you're looking for doesn't exist or has expired" }
                    , viewNoteNotFound
                    ]

        NotFound ->
            [ viewNoteNotFound ]



-- HEADER


viewHeader : { title : String, subtitle : String } -> Html msg
viewHeader options =
    H.div [ A.class "p-6 pb-4 border-b border-gray-200" ]
        [ H.h1
            [ A.class "text-2xl font-bold text-gray-900" ]
            [ H.text options.title ]
        , H.p [ A.class "text-gray-600 mt-2" ] [ H.text options.subtitle ]
        ]


viewShowNoteHeader : Zone -> String -> Note -> Html Msg
viewShowNoteHeader zone slug note =
    H.div []
        [ Components.Utils.viewIf note.burnBeforeExpiration
            (H.div [ A.class "bg-orange-50 border-b border-orange-200 p-4" ]
                [ H.div [ A.class "flex items-center gap-3" ]
                    [ H.div [ A.class "w-6 h-6 bg-orange-100 rounded-full flex items-center justify-center flex-shrink-0" ]
                        [ Icon.view Icon.Warning "w-4 h-4 text-orange-600" ]
                    , H.p [ A.class "text-orange-800 text-sm font-medium" ]
                        [ H.text "This note was destroyed. If you need to keep it, copy it before closing this window." ]
                    ]
                ]
            )
        , H.div [ A.class "p-6 pb-4 border-b border-gray-200" ]
            [ H.div [ A.class "flex justify-between items-start" ]
                [ H.div []
                    [ H.h1 [ A.class "text-2xl font-bold text-gray-900" ] [ H.text ("Note: " ++ slug) ]
                    , H.div [ A.class "text-sm text-gray-500 mt-2 space-y-1" ]
                        [ H.p [] [ H.text ("Created: " ++ T.toString zone note.createdAt) ]
                        , Components.Utils.viewMaybe note.expiresAt (\n -> H.p [] [ H.text ("Expires at: " ++ T.toString zone n) ])
                        ]
                    ]
                , H.div [ A.class "flex gap-2" ]
                    [ H.button
                        [ E.onClick UserClickedCopyContent
                        , A.class "px-3 py-2 text-sm border border-gray-300 text-gray-700 rounded-md hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-black focus:ring-offset-2 transition-colors"
                        ]
                        [ H.text "Copy Content" ]
                    ]
                ]
            ]
        ]



-- NOTE


viewNoteNotFound : Html msg
viewNoteNotFound =
    H.div [ A.class "p-6" ]
        [ H.div [ A.class "text-center py-12" ]
            [ H.div [ A.class "w-16 h-16 bg-red-100 rounded-full flex items-center justify-center mx-auto mb-4" ]
                [ Icon.view Icon.NotFound "w-8 h-8 text-red-500" ]
            , H.h2 [ A.class "text-xl font-semibold text-gray-900 mb-2" ]
                [ H.text "Note not found" ]
            ]
        ]


viewOpenNote : { slug : String, hasPassword : Bool, isLoading : Bool, password : Maybe String } -> Html Msg
viewOpenNote opts =
    let
        isDisabled =
            opts.hasPassword && Maybe.withDefault "" opts.password == ""

        buttonData =
            let
                base =
                    "px-6 py-3 rounded-md focus:outline-none focus:ring-2 focus:ring-black focus:ring-offset-2 transition-colors"
            in
            if opts.isLoading then
                { text = "Loading Note...", class = base ++ " bg-gray-300 text-gray-500 cursor-not-allowed" }

            else if isDisabled then
                { text = "View Note", class = base ++ " bg-gray-300 text-gray-500 cursor-not-allowed" }

            else
                { text = "View Note", class = base ++ " bg-black text-white hover:bg-gray-800" }
    in
    H.div [ A.class "p-6" ]
        [ H.div [ A.class "text-center py-12" ]
            [ H.div [ A.class "mb-6" ]
                [ H.div [ A.class "w-16 h-16 bg-gray-100 rounded-full flex items-center justify-center mx-auto mb-4" ]
                    [ Icon.view Icon.NoteIcon "w-8 h-8 text-gray-400" ]
                , H.h2 [ A.class "text-lg font-semibold text-gray-900 mb-2" ] [ H.text opts.slug ]
                , H.p [ A.class "text-gray-600 mb-6" ] [ H.text "You're about read and destroy the note." ]
                ]
            , H.form
                [ E.onSubmit UserClickedViewNote
                , A.class "max-w-sm mx-auto space-y-4"
                ]
                [ Components.Utils.viewIf opts.hasPassword
                    (H.div
                        [ A.class "space-y-2" ]
                        [ H.label
                            [ A.class "block text-sm font-medium text-gray-700 text-left" ]
                            [ H.text "Password" ]
                        , H.input
                            [ E.onInput UserUpdatedPassword
                            , A.class "w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-black focus:border-transparent"
                            ]
                            []
                        ]
                    )
                , H.button
                    [ A.class buttonData.class
                    , A.type_ "submit"
                    , A.disabled isDisabled
                    ]
                    [ H.text buttonData.text ]
                ]
            ]
        ]


viewNoteContent : Note -> Html msg
viewNoteContent note =
    H.div [ A.class "p-6" ]
        [ H.div [ A.class "bg-gray-50 border border-gray-200 rounded-md p-4" ]
            [ H.pre
                [ A.class "whitespace-pre-wrap font-mono text-sm text-gray-800 overflow-x-auto" ]
                [ H.text note.content ]
            ]
        ]
