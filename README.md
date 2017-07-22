
[![Free music](http://img.auboutdufil.com/logo32.png)][Au Bout du Fil]

Au Bout du Fil unofficial API
=============================

Unofficial API endpoints for the [Au Bout du Fil] website using go.

A running server is available at https://auboutdufil.shywim.fr

## Usage

    make abdf
    ./abdf [-p port]

## Response format

The server answers with JSON data with the following model:

 - `title`: Title of the audio track
 - `artist`: Artist of the audio track
 - `genres`: Genres of the audio track as array
 - `license`: Audio track's license
 - `cover_art_url`: URL of the cover art
 - `download_url`: Download URL
 - `track_url`: Track's details url
 - `downloads`: Download count
 - `play_count`: Play count
 - `rating`: Users rating for the track
 - `published_date`: Date of publication on the website

## Endpoints

Each query give a response with (at most) 6 musics. You can add a `page` query param to get more musics, starting at 1.

### `/latest`

Provides the last musics published.

### `/best`

Provides the musics with the highest rating.

### `/downloads`

Provides the most downloaded musics.

### `/plays`

Provides the most played musics.

### Options

For each of theses endpoints you can add the following paths (in any order):

    /license/:license

Filter by license. These license values will be converted to uppercase if you pass them as lowercase: 

 - Licence art libre (`ART-LIBRE`)
 - Creative Commons Attribution (`CC-BY`)
 - Creative Commons Attribution-Non Commercial (`CC-BYNC`)
 - Creative Commons Attribution-Non Commercial-No Derivative (`CC-BYNCND`)
 - Creative Commons Attribution-Non Commercial-Share Alike (`CC-BYNCSA`)
 - Creative Commons Attribution-No Derivative (`CC-BYND`)
 - Creative Commons Attribution-Share Alike (`CC-BYSA`)
 - Creative Commons Public Domain (`CC0`)

⚠️ For forward compatibility, no error will be sent if an unsupported license is passed and the scrapper will try to fetch music for the value it received. Results in this case may not be accurate.

    /mood/:mood

Filter by mood. You can passe the french mood (as displayed on the [Au Bout du Fil] sidebar) and the corresponding variable will be used:

 - *rageuse*: `angry`
 - *lumineuse*: `bright`
 - *calme*: `calm`
 - *lugubre*: `dark`
 - *dramatique*: `dramatic`
 - *euphorique*: `funky`
 - *heureuse*: `happy`
 - *inspirante*: `inspirational`
 - *romatique*: `romantic`
 - *triste*: `sad`

⚠️ For forward compatibility, no error will be sent if an unknown mood is passed and the scrapper will try to fetch music for the value it received. Results in this case may not be accurate.

    /genre/:genre

Filter by genre. Values passed will be used directly by the scrapper. You can find the list in the [Au Bout du Fil] sidebar.

Genre, mood and license can also be passed as query params to the endpoint (e.g. `/latest?license=cc-by`). If you pass one of these both as a query param and as a url path, the url path will have priority over the query param.


 [Au Bout du Fil]: http://www.auboutdufil.com
