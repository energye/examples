#ifndef __GLIB_GO_H__
#define __GLIB_GO_H__

#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>

#include <gio/gio.h>
#define G_SETTINGS_ENABLE_BACKEND
#include <gio/gsettingsbackend.h>
#include <glib-object.h>
#include <glib.h>
#include <glib/gi18n.h>
#include <locale.h>


static GIcon *toGIcon(void *p) { return (G_ICON(p)); }
static GFileIcon *toGFileIcon(void *p) { return (G_FILE_ICON(p)); }

static GFile *toGFile(void *p) { return (G_FILE(p)); }


#endif
