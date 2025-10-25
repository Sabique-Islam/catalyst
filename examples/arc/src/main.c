#include <stdio.h>
#include <sqlite3.h>

// Callback function for SELECT queries
int callback(void *NotUsed, int argc, char **argv, char **azColName) {
    (void)NotUsed; // suppress unused warning
    for (int i = 0; i < argc; i++) {
        printf("%s = %s\n", azColName[i], argv[i] ? argv[i] : "NULL");
    }
    printf("\n");
    return 0;
}

int main() {
    sqlite3 *db;
    char *errMsg = 0;
    int rc;

    // Open (or create) the database file
    rc = sqlite3_open("test.db", &db);
    if (rc) {
        fprintf(stderr, "Can't open database: %s\n", sqlite3_errmsg(db));
        return 1;
    } else {
        printf("Opened database successfully!\n");
    }

    // Create a table
    const char *create_sql = 
        "CREATE TABLE IF NOT EXISTS USERS("
        "ID INTEGER PRIMARY KEY AUTOINCREMENT,"
        "NAME TEXT NOT NULL,"
        "AGE INT);";
    
    rc = sqlite3_exec(db, create_sql, 0, 0, &errMsg);
    if (rc != SQLITE_OK) {
        fprintf(stderr, "SQL error: %s\n", errMsg);
        sqlite3_free(errMsg);
    } else {
        printf("Table created successfully!\n");
    }

    // Insert some data
    const char *insert_sql = 
        "INSERT INTO USERS (NAME, AGE) VALUES ('Alice', 22);";
    rc = sqlite3_exec(db, insert_sql, 0, 0, &errMsg);
    if (rc != SQLITE_OK) {
        fprintf(stderr, "SQL error: %s\n", errMsg);
        sqlite3_free(errMsg);
    } else {
        printf("Record inserted successfully!\n");
    }

    // Read and display the data
    const char *select_sql = "SELECT * FROM USERS;";
    printf("\n--- USER DATA ---\n");
    rc = sqlite3_exec(db, select_sql, callback, 0, &errMsg);
    if (rc != SQLITE_OK) {
        fprintf(stderr, "SQL error: %s\n", errMsg);
        sqlite3_free(errMsg);
    }

    sqlite3_close(db);
    printf("Database closed.\n");
    return 0;
}