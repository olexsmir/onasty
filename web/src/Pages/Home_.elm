module Pages.Home_ exposing (Model, Msg, PageVariant, page)

import Api
import Api.Note
import Components.Error
import Data.Note as Note
import Effect exposing (Effect)
import Html as H exposing (Html)
import Html.Attributes as A
import Html.Events as E
import Layouts
import Page exposing (Page)
import Process
import Route exposing (Route)
import Shared
import Task
import View exposing (View)


page : Shared.Model -> Route () -> Page Model Msg
page shared _ =
    Page.new
        { init = init shared
        , update = update shared
        , subscriptions = subscriptions
        , view = view shared
        }
        |> Page.withLayout (\_ -> Layouts.Header {})



-- INIT


type alias Model =
    { pageVariant : PageVariant
    , content : String
    , slug : Maybe String
    , password : Maybe String
    , apiError : Maybe Api.Error
    , userClickedCopyLink : Bool
    }



-- TODO: store slug as Slug type


type PageVariant
    = CreateNote
    | NoteCreated String


init : Shared.Model -> () -> ( Model, Effect Msg )
init _ () =
    ( { pageVariant = CreateNote
      , content = ""
      , slug = Nothing
      , password = Nothing
      , userClickedCopyLink = False
      , apiError = Nothing
      }
    , Effect.none
    )



-- UPDATE


type Msg
    = CopyButtonReset
    | UserUpdatedInput Field String
    | UserClickedSubmit
    | UserClickedCreateNewNote
    | UserClickedCopyLink
    | ApiCreateNoteResponded (Result Api.Error Note.CreateResponse)


type Field
    = Content
    | Slug
    | Password


update : Shared.Model -> Msg -> Model -> ( Model, Effect Msg )
update shared msg model =
    case msg of
        CopyButtonReset ->
            ( { model | userClickedCopyLink = False }, Effect.none )

        UserClickedSubmit ->
            ( model
            , Api.Note.create
                { onResponse = ApiCreateNoteResponded
                , content = model.content
                , slug = model.slug
                , password = model.password
                }
            )

        UserClickedCreateNewNote ->
            ( { model
                | pageVariant = CreateNote
                , content = ""
                , slug = Nothing
                , password = Nothing
                , apiError = Nothing
              }
            , Effect.none
            )

        UserClickedCopyLink ->
            ( { model | userClickedCopyLink = True }
            , Effect.batch
                [ Effect.sendCmd (Task.perform (\_ -> CopyButtonReset) (Process.sleep 2000))
                , Effect.sendToClipboard (secretUrl shared.appURL (Maybe.withDefault "" model.slug))
                ]
            )

        UserUpdatedInput Content content ->
            ( { model | content = content }, Effect.none )

        UserUpdatedInput Slug slug ->
            ( { model | slug = Just slug }, Effect.none )

        UserUpdatedInput Password password ->
            ( { model | password = Just password }, Effect.none )

        ApiCreateNoteResponded (Ok response) ->
            ( { model | pageVariant = NoteCreated response.slug, slug = Just response.slug }, Effect.none )

        ApiCreateNoteResponded (Err error) ->
            ( { model | apiError = Just error }, Effect.none )



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions _ =
    Sub.none



-- VIEW


secretUrl : String -> String -> String
secretUrl appUrl slug =
    appUrl ++ "/secret/" ++ slug


view : Shared.Model -> Model -> View Msg
view shared model =
    { title = "Onasty"
    , body =
        [ H.div [ A.class "py-8 px-4 " ]
            [ H.div [ A.class "w-full max-w-4xl mx-auto" ]
                [ H.div [ A.class "bg-white rounded-lg border border-gray-200 shadow-sm" ]
                    [ viewHeader model.pageVariant
                    , H.div [ A.class "p-6 space-y-6" ]
                        [ case model.apiError of
                            Just error ->
                                Components.Error.error (Api.errorMessage error)

                            Nothing ->
                                H.text ""
                        , case model.pageVariant of
                            CreateNote ->
                                viewCreateNoteForm model

                            NoteCreated slug ->
                                viewNoteCreated model.userClickedCopyLink shared.appURL slug
                        ]
                    ]
                ]
            ]
        ]
    }


viewHeader : PageVariant -> Html Msg
viewHeader pageVariant =
    H.div [ A.class "p-6 pb-4 border-b border-gray-200" ]
        [ H.h1 [ A.class "text-2xl font-bold text-gray-900" ]
            [ H.text
                (case pageVariant of
                    CreateNote ->
                        "Create a new note"

                    NoteCreated _ ->
                        "Paste Created Successfully!"
                )
            ]
        ]



-- VIEW CREATE NOTE
-- TODO: validate form


viewCreateNoteForm : Model -> String -> Html Msg
viewCreateNoteForm model appUrl =
    H.form
        [ E.onSubmit UserClickedSubmit
        , A.class "space-y-6"
        ]
        [ viewTextarea
        , viewFormInput
            { field = Slug
            , label = "Custom URL Slug (optional)"
            , placeholder = "my-unique-slug"
            , type_ = "text"
            , help = Just "Leave empty to generate a random slug"
            , prefix = Just (secretUrl appUrl "")
            }
        , H.div [ A.class "flex justify-end" ] [ viewSubmitButton model ]
        ]


viewTextarea : Html Msg
viewTextarea =
    H.div [ A.class "space-y-2" ]
        [ H.label
            [ A.for (fromFieldToName Content)
            , A.class "block text-sm font-medium text-gray-700 mb-2"
            ]
            [ H.text "Content" ]
        , H.textarea
            [ E.onInput (UserUpdatedInput Content)
            , A.id (fromFieldToName Content)
            , A.placeholder "Write your note here..."
            , A.required True
            , A.rows 20
            , A.class "w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-black focus:border-transparent resize-vertical font-mono text-sm"
            ]
            []
        ]


viewFormInput : { field : Field, label : String, placeholder : String, type_ : String, prefix : Maybe String, help : Maybe String } -> Html Msg
viewFormInput options =
    H.div [ A.class "space-y-2" ]
        [ H.label
            [ A.for (fromFieldToName options.field)
            , A.class "block text-sm font-medium text-gray-700 mb-2"
            ]
            [ H.text options.label ]
        , H.div [ A.class "flex items-center" ]
            [ case options.prefix of
                Just prefix ->
                    H.span [ A.class "text-gray-500 text-md mr-2 whitespace-nowrap" ] [ H.text prefix ]

                Nothing ->
                    H.text ""
            , H.input
                [ E.onInput (UserUpdatedInput options.field)
                , A.id (fromFieldToName options.field)
                , A.type_ options.type_
                , A.placeholder options.placeholder
                , A.class "w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-black focus:border-transparent"
                ]
                []
            ]
        , case options.help of
            Just help ->
                H.p [ A.class "text-xs text-gray-500 mt-1" ] [ H.text help ]

            Nothing ->
                H.text ""
        ]


viewSubmitButton : Model -> Html Msg
viewSubmitButton model =
    H.button
        [ A.type_ "submit"
        , A.disabled (isFormDisabled model)
        , A.class
            (if isFormDisabled model then
                "px-6 py-2 bg-gray-300 text-gray-500 rounded-md cursor-not-allowed transition-colors"

             else
                "px-6 py-2 bg-black text-white rounded-md hover:bg-gray-800 focus:outline-none focus:ring-2 focus:ring-black focus:ring-offset-2 transition-colors"
            )
        ]
        [ H.text "Create note" ]


isFormDisabled : Model -> Bool
isFormDisabled model =
    String.isEmpty model.content


viewCreateNewNoteButton : Html Msg
viewCreateNewNoteButton =
    H.button
        [ A.class "px-6 py-2 bg-black text-white rounded-md hover:bg-gray-800 focus:outline-none focus:ring-2 focus:ring-black focus:ring-offset-2 transition-colors"
        , E.onClick UserClickedCreateNewNote
        ]
        [ H.text "Create New Paste" ]


fromFieldToName : Field -> String
fromFieldToName field =
    case field of
        Content ->
            "content"

        Slug ->
            "slug"

        Password ->
            "password"



-- VIEW NOTE CREATED


viewNoteCreated : Bool -> String -> String -> Html Msg
viewNoteCreated userClickedCopyLink appUrl slug =
    H.div [ A.class "bg-green-50 border border-green-200 rounded-md p-6" ]
        [ H.div [ A.class "bg-white border border-green-300 rounded-md p-4 mb-4" ]
            [ H.p [ A.class "text-sm text-gray-600 mb-2" ]
                [ H.text "Your paste is available at:" ]
            , H.p [ A.class "font-mono text-sm text-gray-800 break-all" ]
                [ H.text (secretUrl appUrl slug) ]
            ]
        , H.div [ A.class "flex gap-3" ]
            [ viewCopyLinkButton userClickedCopyLink
            , viewCreateNewNoteButton
            ]
        ]


viewCopyLinkButton : Bool -> Html Msg
viewCopyLinkButton isClicked =
    let
        base : String
        base =
            "px-4 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-black focus:ring-offset-2 transition-colors"
    in
    H.button
        [ A.class
            (if isClicked then
                base ++ " bg-green-100 border-green-300 text-green-700"

             else
                base ++ " border-gray-300 text-gray-700 hover:bg-gray-50"
            )
        , E.onClick UserClickedCopyLink
        ]
        [ H.text
            (if isClicked then
                "Copied!"

             else
                "Copy URL"
            )
        ]
