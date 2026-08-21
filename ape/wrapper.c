/*
 * Cosmopolitan APE wrapper for the file-native CRM.
 * The Go runtime is not compiled by cosmocc. This APE embeds a native
 * host blob (zip member) and execs it.
 *
 * Implements: SYS-REQ-260821-AFPN SW-REQ-260821-AC3S
 */
#include <errno.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/stat.h>
#include <unistd.h>

static int copy_stream(FILE *in, FILE *out) {
  char buf[8192];
  size_t n;
  while ((n = fread(buf, 1, sizeof buf, in)) > 0) {
    if (fwrite(buf, 1, n, out) != n) return -1;
  }
  return ferror(in) ? -1 : 0;
}

static int extract(const char *src, const char *dst) {
  FILE *in = fopen(src, "rb");
  if (!in) return -1;
  FILE *out = fopen(dst, "wb");
  if (!out) {
    fclose(in);
    return -1;
  }
  int rc = copy_stream(in, out);
  fclose(in);
  if (fclose(out) != 0) rc = -1;
  if (rc == 0) chmod(dst, 0700);
  return rc;
}

static void pick_blob(char *path, size_t n) {
  const char *os = "unknown";
  const char *arch = "unknown";
#ifdef __APPLE__
  os = "darwin";
#elif defined(__linux__)
  os = "linux";
#elif defined(_WIN32)
  os = "windows";
#endif
#if defined(__aarch64__) || defined(__arm64__)
  arch = "arm64";
#elif defined(__x86_64__) || defined(__amd64__)
  arch = "x86_64";
#endif
  snprintf(path, n, "/zip/crm.%s-%s", os, arch);
}

int main(int argc, char **argv) {
  char blob[256];
  pick_blob(blob, sizeof blob);
  FILE *probe = fopen(blob, "rb");
  if (!probe) {
    snprintf(blob, sizeof blob, "/zip/crm");
    probe = fopen(blob, "rb");
  }
  if (probe) fclose(probe);
  else {
    fprintf(stderr, "crm.com: no native blob for this host (looked for OS/arch zip member)\n");
    return 127;
  }

  const char *home = getenv("HOME");
  char dest[1024];
  if (home && home[0]) {
    snprintf(dest, sizeof dest, "%s/.cache/crm", home);
    mkdir(dest, 0755);
    snprintf(dest, sizeof dest, "%s/.cache/crm/crm-native", home);
  } else {
    snprintf(dest, sizeof dest, "/tmp/crm-native-%d", (int)getpid());
  }
  if (extract(blob, dest) != 0) {
    fprintf(stderr, "crm.com: extract %s -> %s failed: %s\n", blob, dest, strerror(errno));
    return 126;
  }
  execv(dest, argv);
  fprintf(stderr, "crm.com: exec %s failed: %s\n", dest, strerror(errno));
  return 126;
}
