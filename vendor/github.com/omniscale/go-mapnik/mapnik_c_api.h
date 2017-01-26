#include "stdint.h"
#ifndef MAPNIK_C_API_H
#define MAPNIK_C_API_H

#include <mapnik/version.hpp>

#if defined(WIN32) || defined(WINDOWS) || defined(_WIN32) || defined(_WINDOWS)
#  define MAPNIKCAPICALL __declspec(dllexport)
#else
#  define MAPNIKCAPICALL
#endif

#ifdef __cplusplus
extern "C"
{
#endif

const int mapnik_version = MAPNIK_VERSION;
const int mapnik_version_major = MAPNIK_MAJOR_VERSION;
const int mapnik_version_minor = MAPNIK_MINOR_VERSION;
const int mapnik_version_patch = MAPNIK_PATCH_VERSION;

MAPNIKCAPICALL int mapnik_register_datasources(const char* path);
MAPNIKCAPICALL int mapnik_register_fonts(const char* path);

const int MAPNIK_NONE = 0;
const int MAPNIK_DEBUG = 1;
const int MAPNIK_WARN = 2;
const int MAPNIK_ERROR = 3;

MAPNIKCAPICALL void mapnik_logging_set_severity(int);

MAPNIKCAPICALL const char * mapnik_register_last_error();

// BBOX
typedef struct _mapnik_bbox_t mapnik_bbox_t;
MAPNIKCAPICALL mapnik_bbox_t * mapnik_bbox(double minx, double miny, double maxx, double maxy);
MAPNIKCAPICALL void mapnik_bbox_free(mapnik_bbox_t * b);


// Image
MAPNIKCAPICALL typedef struct _mapnik_image_t mapnik_image_t;
MAPNIKCAPICALL void mapnik_image_free(mapnik_image_t * i);

MAPNIKCAPICALL const char * mapnik_image_last_error(mapnik_image_t * i);

typedef struct _mapnik_image_blob_t {
    char *ptr;
    unsigned int len;
} mapnik_image_blob_t;
MAPNIKCAPICALL void mapnik_image_blob_free(mapnik_image_blob_t * b);

MAPNIKCAPICALL mapnik_image_blob_t * mapnik_image_to_blob(mapnik_image_t * i, const char * format);

MAPNIKCAPICALL const uint8_t * mapnik_image_to_raw(mapnik_image_t * i, size_t *size);
MAPNIKCAPICALL mapnik_image_t * mapnik_image_from_raw(const uint8_t * raw, int width, int height);

//  Map
typedef struct _mapnik_map_t mapnik_map_t;

MAPNIKCAPICALL mapnik_map_t * mapnik_map( unsigned int width, unsigned int height );
MAPNIKCAPICALL void mapnik_map_free(mapnik_map_t * m);

MAPNIKCAPICALL const char * mapnik_map_last_error(mapnik_map_t * m);

MAPNIKCAPICALL int mapnik_map_load(mapnik_map_t * m, const char* stylesheet);

MAPNIKCAPICALL const char * mapnik_map_get_srs(mapnik_map_t * m);
MAPNIKCAPICALL int mapnik_map_set_srs(mapnik_map_t * m, const char* srs);
MAPNIKCAPICALL void mapnik_map_resize(mapnik_map_t * m, unsigned int width, unsigned int height);
MAPNIKCAPICALL double mapnik_map_get_scale_denominator(mapnik_map_t * m);
MAPNIKCAPICALL void mapnik_map_set_buffer_size(mapnik_map_t * m, int buffer_size);

MAPNIKCAPICALL int mapnik_map_background(mapnik_map_t * m, uint8_t *r, uint8_t *g, uint8_t *b, uint8_t *a);
MAPNIKCAPICALL void mapnik_map_set_background(mapnik_map_t * m, uint8_t r, uint8_t g, uint8_t b, uint8_t a);

MAPNIKCAPICALL int mapnik_map_zoom_all(mapnik_map_t * m);
MAPNIKCAPICALL void mapnik_map_zoom_to_box(mapnik_map_t * m, mapnik_bbox_t * b);

MAPNIKCAPICALL void mapnik_map_set_maximum_extent(mapnik_map_t * m, double x0, double y0, double x1, double y1);
MAPNIKCAPICALL void mapnik_map_reset_maximum_extent(mapnik_map_t * m);

MAPNIKCAPICALL int mapnik_map_render_to_file(mapnik_map_t * m, const char* filepath, double scale, double scale_factor, const char *format);
MAPNIKCAPICALL mapnik_image_t * mapnik_map_render_to_image(mapnik_map_t * m, double scale, double scale_factor);

MAPNIKCAPICALL int mapnik_map_layer_count(mapnik_map_t * m);
MAPNIKCAPICALL const char * mapnik_map_layer_name(mapnik_map_t * m, size_t idx);
MAPNIKCAPICALL int mapnik_map_layer_is_active(mapnik_map_t * m, size_t idx);
MAPNIKCAPICALL void mapnik_map_layer_set_active(mapnik_map_t * m, size_t idx, int active);

#ifdef __cplusplus
}
#endif

#endif // MAPNIK_C_API_H
