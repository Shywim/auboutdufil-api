
[![Free music](http://img.auboutdufil.com/logo32.png)][abdf]

Au Bout du Fil unnofficial API
=============================

Unnofficial API endpoints for the [Au Bout du Fil][abdf] website using go.

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

### /latest

Provides the last 6 tracks published.

 [abdf]: http://www.auboutdufil.com
