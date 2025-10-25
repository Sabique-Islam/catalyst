#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <curl/curl.h>
#include <jansson.h>

struct MemoryStruct {
    char *memory;
    size_t size;
};

// Callback for libcurl to write response into memory
static size_t WriteMemoryCallback(void *contents, size_t size, size_t nmemb, void *userp) {
    size_t realsize = size * nmemb;
    struct MemoryStruct *mem = (struct MemoryStruct *)userp;

    char *ptr = realloc(mem->memory, mem->size + realsize + 1);
    if(ptr == NULL) {
        printf("Not enough memory (realloc returned NULL)\n");
        return 0;
    }

    mem->memory = ptr;
    memcpy(&(mem->memory[mem->size]), contents, realsize);
    mem->size += realsize;
    mem->memory[mem->size] = 0;

    return realsize;
}

void search_papers(const char *keyword) {
    CURL *curl;
    CURLcode res;

    struct MemoryStruct chunk;
    chunk.memory = malloc(1);
    chunk.size = 0;

    curl_global_init(CURL_GLOBAL_DEFAULT);
    curl = curl_easy_init();
    if(curl) {
        // URL-encode the keyword
        char *encoded_keyword = curl_easy_escape(curl, keyword, 0);

        char url[1024];
        snprintf(url, sizeof(url),
            "https://api.semanticscholar.org/graph/v1/paper/search?query=%s&fields=title&limit=5",
            encoded_keyword);

        curl_easy_setopt(curl, CURLOPT_URL, url);
        curl_easy_setopt(curl, CURLOPT_WRITEFUNCTION, WriteMemoryCallback);
        curl_easy_setopt(curl, CURLOPT_WRITEDATA, (void *)&chunk);

        res = curl_easy_perform(curl);

        if(res != CURLE_OK) {
            fprintf(stderr, "curl_easy_perform() failed: %s\n", curl_easy_strerror(res));
        } else {
            json_error_t error;
            json_t *root = json_loads(chunk.memory, 0, &error);
            if(!root) {
                fprintf(stderr, "JSON error on line %d: %s\n", error.line, error.text);
            } else {
                json_t *data = json_object_get(root, "data");
                if(json_is_array(data)) {
                    size_t index;
                    json_t *paper;
                    printf("\nTop %zu results for keyword \"%s\":\n", json_array_size(data), keyword);
                    json_array_foreach(data, index, paper) {
                        json_t *title = json_object_get(paper, "title");
                        if(json_is_string(title)) {
                            printf("%zu. %s\n", index + 1, json_string_value(title));
                        }
                    }
                } else {
                    printf("No results found.\n");
                }
                json_decref(root);
            }
        }

        curl_easy_cleanup(curl);
        curl_free(encoded_keyword);
    }

    free(chunk.memory);
    curl_global_cleanup();
}

int main(int argc, char *argv[]) {
    char keyword[256];

    // Check if keyword was provided as command-line argument
    if(argc > 1) {
        // Concatenate all arguments into one search string
        keyword[0] = '\0';
        for(int i = 1; i < argc; i++) {
            if(i > 1) strcat(keyword, " ");
            strncat(keyword, argv[i], sizeof(keyword) - strlen(keyword) - 1);
        }
        
        printf("Searching for: %s\n", keyword);
        search_papers(keyword);
    } else {
        // Interactive mode - prompt user for input
        printf("Enter keyword to search for research papers: ");
        if(fgets(keyword, sizeof(keyword), stdin)) {
            // Remove trailing newline if present
            size_t len = strlen(keyword);
            if(len > 0 && keyword[len - 1] == '\n') {
                keyword[len - 1] = '\0';
            }

            if(strlen(keyword) == 0) {
                printf("No keyword entered. Exiting.\n");
                return 1;
            }

            search_papers(keyword);
        } else {
            printf("Failed to read input.\n");
            return 1;
        }
    }

    return 0;
}
