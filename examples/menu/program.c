#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sqlite3.h>

void add_task(sqlite3 *db, const char *task_name) {
    char sql[256];
    snprintf(sql, sizeof(sql), "INSERT INTO tasks (name, done) VALUES ('%s', 0);", task_name);
    char *errMsg = 0;
    if (sqlite3_exec(db, sql, 0, 0, &errMsg) != SQLITE_OK) {
        fprintf(stderr, "Error adding task: %s\n", errMsg);
        sqlite3_free(errMsg);
    } else {
        printf("Task added: %s\n", task_name);
    }
}

void list_tasks(sqlite3 *db) {
    const char *sql = "SELECT id, name, done FROM tasks;";
    sqlite3_stmt *stmt;
    if (sqlite3_prepare_v2(db, sql, -1, &stmt, 0) != SQLITE_OK) {
        fprintf(stderr, "Failed to fetch tasks\n");
        return;
    }

    printf("ID  | Done | Task\n");
    printf("---------------------\n");
    while (sqlite3_step(stmt) == SQLITE_ROW) {
        int id = sqlite3_column_int(stmt, 0);
        const char *name = (const char*)sqlite3_column_text(stmt, 1);
        int done = sqlite3_column_int(stmt, 2);
        printf("%2d  |  %c   | %s\n", id, done ? 'X' : ' ', name);
    }
    sqlite3_finalize(stmt);
}

void mark_done(sqlite3 *db, int id) {
    char sql[128];
    snprintf(sql, sizeof(sql), "UPDATE tasks SET done = 1 WHERE id = %d;", id);
    char *errMsg = 0;
    if (sqlite3_exec(db, sql, 0, 0, &errMsg) != SQLITE_OK) {
        fprintf(stderr, "Error marking task done: %s\n", errMsg);
        sqlite3_free(errMsg);
    } else {
        printf("Task %d marked as done\n", id);
    }
}

int main() {
    sqlite3 *db;
    if (sqlite3_open("tasks.db", &db)) {
        fprintf(stderr, "Can't open database: %s\n", sqlite3_errmsg(db));
        return 1;
    }

    // Create table if it doesn't exist
    const char *create_sql = "CREATE TABLE IF NOT EXISTS tasks (id INTEGER PRIMARY KEY, name TEXT, done INTEGER);";
    char *errMsg = 0;
    sqlite3_exec(db, create_sql, 0, 0, &errMsg);

    printf("Task Manager\n");
    printf("Commands: add <task>, list, done <id>, exit\n");

    char input[256];
    while (1) {
        printf("> ");
        if (!fgets(input, sizeof(input), stdin)) break;

        // Remove newline
        input[strcspn(input, "\n")] = 0;

        if (strncmp(input, "add ", 4) == 0) {
            add_task(db, input + 4);
        } else if (strcmp(input, "list") == 0) {
            list_tasks(db);
        } else if (strncmp(input, "done ", 5) == 0) {
            int id = atoi(input + 5);
            mark_done(db, id);
        } else if (strcmp(input, "exit") == 0) {
            break;
        } else {
            printf("Unknown command\n");
        }
    }

    sqlite3_close(db);
    return 0;
}
