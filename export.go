package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"

	"github.com/MasterShizzle/goda-stories/gosoh"
	"github.com/davecgh/go-spew/spew"
)

const tileInfoFile = "assets/text/tileInfo.txt"
const itemInfoFile = "assets/text/itemInfo.txt"
const puzzleInfoFile = "assets/text/puzzleInfo.txt"
const mapInfoHtml = "assets/text/mapInfo.html"
const mapInfoText = "assets/text/mapInfo.txt"
const crtrInfoText = "assets/text/creatureInfo.txt"
const htmlStarter string = `<!DOCTYPE html>
<html lang="en">

<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<meta http-equiv="X-UA-Compatible" content="ie=edge">
	<title>Map Info</title>
	<style>
	div {
		float: left;
	}
	img {
		margin: 0px;
	}
	div.mapimg {
		border: 4px outset gray;
		line-height: 0px;
		position: relative;
	}
	div.mapgrid {
		position: absolute;
		top: 0px;
		left: 0px;
		width: 100%;
		height: 100%;
		border: 0px;
		margin: 0px;
		padding: 0px;
		background-image:
		linear-gradient(to right, grey 1px, transparent 1px),
		linear-gradient(to bottom, grey 1px, transparent 1px);
		background-size: 32px 32px;
	}
	div.mapcontainer {
		display: flex;
		width: 95%;
	}
	.textbox {
		padding: 0.5em;
	}
	table {
		width: 30%;
		font-size: 0.5em;
	}
	td {
		width: 30px;
		border-bottom: 1px solid black;
		border-right: 1px solid gray;
	}
	</style>
</head>
<body>
<h1>Yoda Stories Map Info</h1>
`

func dumpToFile(filepath string, foo ...interface{}) error {
	// Create the output file
	file, err := os.Create(filepath)

	spew.Fdump(file, foo)
	fmt.Printf("[dumpToFile] Saved to file: %s\n", filepath)
	return err
}

func getZoneHTML(zone gosoh.ZoneInfo) (ret string) {
	mapName := "map_" + fmt.Sprintf("%03d", zone.Id)
	// For formatting purposes, stop at 5 maps
	// if zone.Id > 5 {
	// 	ret = ""
	// 	return
	// }
	ret = fmt.Sprintf("<div class=\"mapcontainer\"><hr></div>\n<h2 id=\"%s\">%s</h2>\n\n<div class=\"mapcontainer\">\n", mapName, mapName)
	ret += fmt.Sprintf("<div class=\"mapimg\"><img src=\"../maps/%s.png\" alt=\"%s\"><div class=\"mapgrid\"></div></div>\n", mapName, mapName)
	ret += fmt.Sprintf("<p class=\"textbox\">Type: %s, %dx%d (%s)</p>\n\n", zone.Biome, zone.Width, zone.Height, zone.Type)

	ret += "<div class=\"textbox\"><b>Object Triggers</b>\n\n"
	if zone.TileTriggers == nil {
		ret += "<p>None</p>\n</div>\n</div>\n\n"
	} else {
		ret += "<ul>\n"
		for i := 0; i < len(zone.TileTriggers); i++ {
			t := zone.TileTriggers[i]
			// Reminder: we can print info about the tile here, as well
			switch t.Type {
			case "map_entrance":
				ret += "  <li>Map entrance" + fmt.Sprintf(" (%d, %d) => <a href=\"#map_%03d\">map_%03d</a></li>\n", t.X, t.Y, t.Arg, t.Arg)
			case "vehicle_to_secondary_map":
				ret += "  <li>Vehicle" + fmt.Sprintf(" (%d, %d) => <a href=\"#map_%03d\">map_%03d</a></li>\n", t.X, t.Y, t.Arg, t.Arg)
			case "xwing_to_dagobah", "xwing_from_dagobah":
				ret += "  <li>XWing" + fmt.Sprintf(" (%d, %d) => <a href=\"#map_%03d\">map_%03d</a></li>\n", t.X, t.Y, t.Arg, t.Arg)
			default:
				ret += "  <li>" + t.Type + fmt.Sprintf(" (%d, %d) arg %d</li>\n", t.X, t.Y, t.Arg)
			}
		}
		ret += "</ul>\n</div>\n</div>\n<br></br>\n\n"
	}

	ret += "<div class=\"mapcontainer\">\n"
	ret += "<b>Terrain</b>\n\n" + getHTMLTableFromMap(zone.TileMaps.Terrain, zone.Width, zone.Height)
	ret += "<b>Objects</b>\n\n" + getHTMLTableFromMap(zone.TileMaps.Objects, zone.Width, zone.Height)
	ret += "<b>Overlay</b>\n\n" + getHTMLTableFromMap(zone.TileMaps.Overlay, zone.Width, zone.Height) + "\n</div>\n<br></br>\n\n"

	return
}

func getHTMLTableFromMap(zone []int, zWidth, zHeight int) (ret string) {
	// Convert a map layer into an HTML table with tile numbers
	ret = "<table>\n  <tr><th></th>"
	for i := 0; i < zWidth; i++ {
		ret += fmt.Sprintf("<td><b>%02d</b></td>", i)
	}
	ret += "</tr>\n"

	// The March of the Ints!
	for j := 0; j < zHeight; j++ {
		ret += fmt.Sprintf("  <tr><td><b>%02d</b></td>", j)
		for i := 0; i < zWidth; i++ {
			// TODO: maybe lookup tile types and apply some colors to each cell, by type, or with CSS
			tileNum := zone[i+(zWidth*j)]
			if tileNum == 65535 {
				// Show transparent tiles as blank
				ret += "<td></td>"
			} else {
				ret += fmt.Sprintf("<td>%04d</td>", tileNum)
			}
		}
		ret += "</tr>\n"
	}
	ret += "</table>\n\n"

	return
}

func getTileByNumber(tileNum int) (image.Image, error) {
	// Get the tile from its .png image source, by number
	blankTile := image.NewRGBA(image.Rect(0, 0, gosoh.TileWidth, gosoh.TileHeight))
	draw.Draw(blankTile, image.Rect(0, 0, gosoh.TileWidth, gosoh.TileHeight), image.Transparent, image.Pt(0, 0), draw.Src)
	filePath := "assets/tiles/tile_" + fmt.Sprintf("%04d", tileNum) + ".png"
	if tileNum == 65535 { // return a transparent tile
		return blankTile, nil
	}
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	image, _, err := image.Decode(f)
	return image, err
}

func saveMapToPNG(mapPath string, zone gosoh.ZoneInfo) error {
	// Make a blank map and fill with black
	mapImage := image.NewRGBA(image.Rect(0, 0, zone.Width*gosoh.TileWidth, zone.Height*gosoh.TileHeight))
	draw.Draw(mapImage, mapImage.Bounds(), image.Black, image.Black.Bounds().Max, draw.Src)

	// Draw tiles
	for i := 0; i < (zone.Width * zone.Height); i++ {
		terrainTile, err := getTileByNumber(zone.TileMaps.Terrain[i])
		if err != nil {
			log.Fatal(err)
		}

		objectsTile, err := getTileByNumber(zone.TileMaps.Objects[i])
		if err != nil {
			log.Fatal(err)
		}

		overlayTile, err := getTileByNumber(zone.TileMaps.Overlay[i])
		if err != nil {
			log.Fatal(err)
		}

		x := (i % zone.Width) * gosoh.TileWidth
		y := (i / zone.Height) * gosoh.TileHeight

		offset := image.Pt(x, y)

		if terrainTile != nil {
			draw.Draw(mapImage, mapImage.Bounds().Add(offset), terrainTile, image.Pt(0, 0), draw.Src)
		}
		if objectsTile != nil {
			draw.Draw(mapImage, mapImage.Bounds().Add(offset), objectsTile, image.Pt(0, 0), draw.Over)
		}
		if overlayTile != nil {
			draw.Draw(mapImage, mapImage.Bounds().Add(offset), overlayTile, image.Pt(0, 0), draw.Over)
		}
	}

	f, err := os.Create(mapPath)
	if err != nil {
		return err
	}
	if err := png.Encode(f, mapImage); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return nil
}

func saveByteSliceToPNG(tPath string, tData []byte) error {
	// Palette data extracted from the de-compiled Yoda Stories binary
	PaletteData := []byte{
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF, 0xFF, 0x8B, 0x00, 0xC3, 0xCF, 0x4B, 0x00,
		0x8B, 0xA3, 0x1B, 0x00, 0x57, 0x77, 0x00, 0x00, 0x8B, 0xA3, 0x1B, 0x00, 0xC3, 0xCF, 0x4B, 0x00,
		0xFB, 0xFB, 0xFB, 0x00, 0xEB, 0xE7, 0xE7, 0x00, 0xDB, 0xD3, 0xD3, 0x00, 0xCB, 0xC3, 0xC3, 0x00,
		0xBB, 0xB3, 0xB3, 0x00, 0xAB, 0xA3, 0xA3, 0x00, 0x9B, 0x8F, 0x8F, 0x00, 0x8B, 0x7F, 0x7F, 0x00,
		0x7B, 0x6F, 0x6F, 0x00, 0x67, 0x5B, 0x5B, 0x00, 0x57, 0x4B, 0x4B, 0x00, 0x47, 0x3B, 0x3B, 0x00,
		0x33, 0x2B, 0x2B, 0x00, 0x23, 0x1B, 0x1B, 0x00, 0x13, 0x0F, 0x0F, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0xC7, 0x43, 0x00, 0x00, 0xB7, 0x43, 0x00, 0x00, 0xAB, 0x3F, 0x00, 0x00, 0x9F, 0x3F, 0x00,
		0x00, 0x93, 0x3F, 0x00, 0x00, 0x87, 0x3B, 0x00, 0x00, 0x7B, 0x37, 0x00, 0x00, 0x6F, 0x33, 0x00,
		0x00, 0x63, 0x33, 0x00, 0x00, 0x53, 0x2B, 0x00, 0x00, 0x47, 0x27, 0x00, 0x00, 0x3B, 0x23, 0x00,
		0x00, 0x2F, 0x1B, 0x00, 0x00, 0x23, 0x13, 0x00, 0x00, 0x17, 0x0F, 0x00, 0x00, 0x0B, 0x07, 0x00,
		0x4B, 0x7B, 0xBB, 0x00, 0x43, 0x73, 0xB3, 0x00, 0x43, 0x6B, 0xAB, 0x00, 0x3B, 0x63, 0xA3, 0x00,
		0x3B, 0x63, 0x9B, 0x00, 0x33, 0x5B, 0x93, 0x00, 0x33, 0x5B, 0x8B, 0x00, 0x2B, 0x53, 0x83, 0x00,
		0x2B, 0x4B, 0x73, 0x00, 0x23, 0x4B, 0x6B, 0x00, 0x23, 0x43, 0x5F, 0x00, 0x1B, 0x3B, 0x53, 0x00,
		0x1B, 0x37, 0x47, 0x00, 0x1B, 0x33, 0x43, 0x00, 0x13, 0x2B, 0x3B, 0x00, 0x0B, 0x23, 0x2B, 0x00,
		0xD7, 0xFF, 0xFF, 0x00, 0xBB, 0xEF, 0xEF, 0x00, 0xA3, 0xDF, 0xDF, 0x00, 0x8B, 0xCF, 0xCF, 0x00,
		0x77, 0xC3, 0xC3, 0x00, 0x63, 0xB3, 0xB3, 0x00, 0x53, 0xA3, 0xA3, 0x00, 0x43, 0x93, 0x93, 0x00,
		0x33, 0x87, 0x87, 0x00, 0x27, 0x77, 0x77, 0x00, 0x1B, 0x67, 0x67, 0x00, 0x13, 0x5B, 0x5B, 0x00,
		0x0B, 0x4B, 0x4B, 0x00, 0x07, 0x3B, 0x3B, 0x00, 0x00, 0x2B, 0x2B, 0x00, 0x00, 0x1F, 0x1F, 0x00,
		0xDB, 0xEB, 0xFB, 0x00, 0xD3, 0xE3, 0xFB, 0x00, 0xC3, 0xDB, 0xFB, 0x00, 0xBB, 0xD3, 0xFB, 0x00,
		0xB3, 0xCB, 0xFB, 0x00, 0xA3, 0xC3, 0xFB, 0x00, 0x9B, 0xBB, 0xFB, 0x00, 0x8F, 0xB7, 0xFB, 0x00,
		0x83, 0xB3, 0xF7, 0x00, 0x73, 0xA7, 0xFB, 0x00, 0x63, 0x9B, 0xFB, 0x00, 0x5B, 0x93, 0xF3, 0x00,
		0x5B, 0x8B, 0xEB, 0x00, 0x53, 0x8B, 0xDB, 0x00, 0x53, 0x83, 0xD3, 0x00, 0x4B, 0x7B, 0xCB, 0x00,
		0x9B, 0xC7, 0xFF, 0x00, 0x8F, 0xB7, 0xF7, 0x00, 0x87, 0xB3, 0xEF, 0x00, 0x7F, 0xA7, 0xF3, 0x00,
		0x73, 0x9F, 0xEF, 0x00, 0x53, 0x83, 0xCF, 0x00, 0x3B, 0x6B, 0xB3, 0x00, 0x2F, 0x5B, 0xA3, 0x00,
		0x23, 0x4F, 0x93, 0x00, 0x1B, 0x43, 0x83, 0x00, 0x13, 0x3B, 0x77, 0x00, 0x0B, 0x2F, 0x67, 0x00,
		0x07, 0x27, 0x57, 0x00, 0x00, 0x1B, 0x47, 0x00, 0x00, 0x13, 0x37, 0x00, 0x00, 0x0F, 0x2B, 0x00,
		0xFB, 0xFB, 0xE7, 0x00, 0xF3, 0xF3, 0xD3, 0x00, 0xEB, 0xE7, 0xC7, 0x00, 0xE3, 0xDF, 0xB7, 0x00,
		0xDB, 0xD7, 0xA7, 0x00, 0xD3, 0xCF, 0x97, 0x00, 0xCB, 0xC7, 0x8B, 0x00, 0xC3, 0xBB, 0x7F, 0x00,
		0xBB, 0xB3, 0x73, 0x00, 0xAF, 0xA7, 0x63, 0x00, 0x9B, 0x93, 0x47, 0x00, 0x87, 0x7B, 0x33, 0x00,
		0x6F, 0x67, 0x1F, 0x00, 0x5B, 0x53, 0x0F, 0x00, 0x47, 0x43, 0x00, 0x00, 0x37, 0x33, 0x00, 0x00,
		0xFF, 0xF7, 0xF7, 0x00, 0xEF, 0xDF, 0xDF, 0x00, 0xDF, 0xC7, 0xC7, 0x00, 0xCF, 0xB3, 0xB3, 0x00,
		0xBF, 0x9F, 0x9F, 0x00, 0xB3, 0x8B, 0x8B, 0x00, 0xA3, 0x7B, 0x7B, 0x00, 0x93, 0x6B, 0x6B, 0x00,
		0x83, 0x57, 0x57, 0x00, 0x73, 0x4B, 0x4B, 0x00, 0x67, 0x3B, 0x3B, 0x00, 0x57, 0x2F, 0x2F, 0x00,
		0x47, 0x27, 0x27, 0x00, 0x37, 0x1B, 0x1B, 0x00, 0x27, 0x13, 0x13, 0x00, 0x1B, 0x0B, 0x0B, 0x00,
		0xF7, 0xB3, 0x37, 0x00, 0xE7, 0x93, 0x07, 0x00, 0xFB, 0x53, 0x0B, 0x00, 0xFB, 0x00, 0x00, 0x00,
		0xCB, 0x00, 0x00, 0x00, 0x9F, 0x00, 0x00, 0x00, 0x6F, 0x00, 0x00, 0x00, 0x43, 0x00, 0x00, 0x00,
		0xBF, 0xBB, 0xFB, 0x00, 0x8F, 0x8B, 0xFB, 0x00, 0x5F, 0x5B, 0xFB, 0x00, 0x93, 0xBB, 0xFF, 0x00,
		0x5F, 0x97, 0xF7, 0x00, 0x3B, 0x7B, 0xEF, 0x00, 0x23, 0x63, 0xC3, 0x00, 0x13, 0x53, 0xB3, 0x00,
		0x00, 0x00, 0xFF, 0x00, 0x00, 0x00, 0xEF, 0x00, 0x00, 0x00, 0xE3, 0x00, 0x00, 0x00, 0xD3, 0x00,
		0x00, 0x00, 0xC3, 0x00, 0x00, 0x00, 0xB7, 0x00, 0x00, 0x00, 0xA7, 0x00, 0x00, 0x00, 0x9B, 0x00,
		0x00, 0x00, 0x8B, 0x00, 0x00, 0x00, 0x7F, 0x00, 0x00, 0x00, 0x6F, 0x00, 0x00, 0x00, 0x63, 0x00,
		0x00, 0x00, 0x53, 0x00, 0x00, 0x00, 0x47, 0x00, 0x00, 0x00, 0x37, 0x00, 0x00, 0x00, 0x2B, 0x00,
		0x00, 0xFF, 0xFF, 0x00, 0x00, 0xE3, 0xF7, 0x00, 0x00, 0xCF, 0xF3, 0x00, 0x00, 0xB7, 0xEF, 0x00,
		0x00, 0xA3, 0xEB, 0x00, 0x00, 0x8B, 0xE7, 0x00, 0x00, 0x77, 0xDF, 0x00, 0x00, 0x63, 0xDB, 0x00,
		0x00, 0x4F, 0xD7, 0x00, 0x00, 0x3F, 0xD3, 0x00, 0x00, 0x2F, 0xCF, 0x00, 0x97, 0xFF, 0xFF, 0x00,
		0x83, 0xDF, 0xEF, 0x00, 0x73, 0xC3, 0xDF, 0x00, 0x5F, 0xA7, 0xCF, 0x00, 0x53, 0x8B, 0xC3, 0x00,
		0x2B, 0x2B, 0x00, 0x00, 0x23, 0x23, 0x00, 0x00, 0x1B, 0x1B, 0x00, 0x00, 0x13, 0x13, 0x00, 0x00,
		0xFF, 0x0B, 0x00, 0x00, 0xFF, 0x00, 0x4B, 0x00, 0xFF, 0x00, 0xA3, 0x00, 0xFF, 0x00, 0xFF, 0x00,
		0x00, 0xFF, 0x00, 0x00, 0x00, 0x4B, 0x00, 0x00, 0xFF, 0xFF, 0x00, 0x00, 0xFF, 0x33, 0x2F, 0x00,
		0x00, 0x00, 0xFF, 0x00, 0x00, 0x1F, 0x97, 0x00, 0xDF, 0x00, 0xFF, 0x00, 0x73, 0x00, 0x77, 0x00,
		0x6B, 0x7B, 0xC3, 0x00, 0x57, 0x57, 0xAB, 0x00, 0x57, 0x47, 0x93, 0x00, 0x53, 0x37, 0x7F, 0x00,
		0x4F, 0x27, 0x67, 0x00, 0x47, 0x1B, 0x4F, 0x00, 0x3B, 0x13, 0x3B, 0x00, 0x27, 0x77, 0x77, 0x00,
		0x23, 0x73, 0x73, 0x00, 0x1F, 0x6F, 0x6F, 0x00, 0x1B, 0x6B, 0x6B, 0x00, 0x1B, 0x67, 0x67, 0x00,
		0x1B, 0x6B, 0x6B, 0x00, 0x1F, 0x6F, 0x6F, 0x00, 0x23, 0x73, 0x73, 0x00, 0x27, 0x77, 0x77, 0x00,
		0xFF, 0xFF, 0xEF, 0x00, 0xF7, 0xF7, 0xDB, 0x00, 0xF3, 0xEF, 0xCB, 0x00, 0xEF, 0xEB, 0xBB, 0x00,
		0xF3, 0xEF, 0xCB, 0x00, 0xE7, 0x93, 0x07, 0x00, 0xE7, 0x97, 0x0F, 0x00, 0xEB, 0x9F, 0x17, 0x00,
		0xEF, 0xA3, 0x23, 0x00, 0xF3, 0xAB, 0x2B, 0x00, 0xF7, 0xB3, 0x37, 0x00, 0xEF, 0xA7, 0x27, 0x00,
		0xEB, 0x9F, 0x1B, 0x00, 0xE7, 0x97, 0x0F, 0x00, 0x0B, 0xCB, 0xFB, 0x00, 0x0B, 0xA3, 0xFB, 0x00,
		0x0B, 0x73, 0xFB, 0x00, 0x0B, 0x4B, 0xFB, 0x00, 0x0B, 0x23, 0xFB, 0x00, 0x0B, 0x73, 0xFB, 0x00,
		0x00, 0x13, 0x93, 0x00, 0x00, 0x0B, 0xD3, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF, 0xFF, 0xFF, 0x00,
	}

	// Make a blank image
	tile := image.NewNRGBA(image.Rect(0, 0, gosoh.TileWidth, gosoh.TileHeight))

	// Set pixels
	for j := 0; j < len(tData); j++ {
		pixel := int(tData[j])
		if pixel == 0 {
			tile.Set(j%gosoh.TileWidth, j/gosoh.TileHeight, color.Transparent)
		} else {
			rVal := PaletteData[pixel*4+2]
			gVal := PaletteData[pixel*4+1]
			bVal := PaletteData[pixel*4+0]
			tile.Set(j%gosoh.TileWidth, j/gosoh.TileHeight, color.NRGBA{
				R: rVal,
				G: gVal,
				B: bVal,
				A: 255,
			})
		}
	}

	// Save tiles to .png
	f, err := os.Create(tPath)
	if err != nil {
		return err
	}
	if err := png.Encode(f, tile); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}

	return nil
}
