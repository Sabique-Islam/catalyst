#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <ncurses.h>
#include <sqlite3.h>
#include <time.h>

#define DB_FILE "bmi_history.db"

// Function to calculate BMI
double calculate_bmi(double weight_kg, double height_m) {
    return weight_kg / (height_m * height_m);
}

// Function to provide advice
const char* bmi_advice(double bmi) {
    if (bmi < 18.5) return "Underweight: Consider healthy diet.";
    else if (bmi < 25.0) return "Normal weight: Keep it up!";
    else if (bmi < 30.0) return "Overweight: Exercise and diet recommended.";
    else return "Obese: Seek medical advice and adopt healthier lifestyle.";
}

// Initialize SQLite database
void init_db(sqlite3 *db) {
    char *err_msg = 0;
    const char *sql = "CREATE TABLE IF NOT EXISTS bmi_records("
                      "id INTEGER PRIMARY KEY, "
                      "name TEXT, "
                      "bmi REAL, "
                      "date TEXT);";

    if (sqlite3_exec(db, sql, 0, 0, &err_msg) != SQLITE_OK) {
        mvprintw(0, 0, "Failed to create table: %s", err_msg);
        sqlite3_free(err_msg);
    }
}

// Insert BMI record into database
void insert_record(sqlite3 *db, const char *name, double bmi) {
    char sql[512];
    time_t t = time(NULL);
    char date_str[64];
    strftime(date_str, sizeof(date_str), "%Y-%m-%d %H:%M:%S", localtime(&t));

    snprintf(sql, sizeof(sql),
             "INSERT INTO bmi_records(name, bmi, date) VALUES('%s', %.2f, '%s');",
             name, bmi, date_str);

    char *err_msg = 0;
    if (sqlite3_exec(db, sql, 0, 0, &err_msg) != SQLITE_OK) {
        mvprintw(0, 0, "Failed to insert record: %s", err_msg);
        sqlite3_free(err_msg);
    }
}

// Display historical BMI records
void view_history(sqlite3 *db) {
    sqlite3_stmt *stmt;
    const char *sql = "SELECT name, bmi, date FROM bmi_records ORDER BY id DESC LIMIT 10;";

    if (sqlite3_prepare_v2(db, sql, -1, &stmt, 0) != SQLITE_OK) {
        mvprintw(0, 0, "Failed to fetch records\n");
        return;
    }

    clear();
    attron(COLOR_PAIR(1) | A_BOLD);
    mvprintw(0, 0, "╔═══════════════════════════════╗");
    mvprintw(1, 0, "║      Last 10 BMI Records      ║");
    mvprintw(2, 0, "╚═══════════════════════════════╝");
    attroff(COLOR_PAIR(1) | A_BOLD);

    int row = 4;
    while (sqlite3_step(stmt) == SQLITE_ROW) {
        const unsigned char *name = sqlite3_column_text(stmt, 0);
        double bmi = sqlite3_column_double(stmt, 1);
        const unsigned char *date = sqlite3_column_text(stmt, 2);
        
        attron(COLOR_PAIR(2));
        mvprintw(row, 2, "Name: %s", name);
        mvprintw(row + 1, 2, "BMI: %.2f", bmi);
        mvprintw(row + 2, 2, "Date: %s", date);
        attroff(COLOR_PAIR(2));
        
        mvprintw(row + 3, 2, "------------------------");
        row += 4;
    }
    
    attron(COLOR_PAIR(3));
    mvprintw(row + 1, 2, "Press any key to return...");
    attroff(COLOR_PAIR(3));
    refresh();
    getch();
    sqlite3_finalize(stmt);
}

// Main UI
int main(void) {
    sqlite3 *db;
    if (sqlite3_open(DB_FILE, &db)) {
        printf("Can't open database: %s\n", sqlite3_errmsg(db));
        return 1;
    }
    init_db(db);

    initscr();
    noecho();
    cbreak();
    start_color();  // Initialize color support
    
    // Define color pairs
    init_pair(1, COLOR_GREEN, COLOR_BLACK);   // For title
    init_pair(2, COLOR_CYAN, COLOR_BLACK);    // For menu items
    init_pair(3, COLOR_YELLOW, COLOR_BLACK);  // For input prompts
    init_pair(4, COLOR_RED, COLOR_BLACK);     // For warnings/important info
    
    while (1) {
        clear();
        attron(COLOR_PAIR(1) | A_BOLD);
        mvprintw(0, 0, "╔═══════════════════════════════╗");
        mvprintw(1, 0, "║   BMI & Nutrition Tracker     ║");
        mvprintw(2, 0, "╚═══════════════════════════════╝");
        attroff(COLOR_PAIR(1) | A_BOLD);

        attron(COLOR_PAIR(2));
        mvprintw(4, 2, "1. Enter new user data");
        mvprintw(5, 2, "2. View last 10 BMI records");
        mvprintw(6, 2, "3. Exit");
        attroff(COLOR_PAIR(2));

        attron(COLOR_PAIR(3));
        mvprintw(8, 0, "Select an option: ");
        attroff(COLOR_PAIR(3));
        refresh();

        int choice = getch() - '0';

        if (choice == 1) {
            char name[50];
            double height, weight;

            clear();
            echo();
            attron(COLOR_PAIR(3));
            mvprintw(2, 2, "Enter name: ");
            attroff(COLOR_PAIR(3));
            getnstr(name, sizeof(name)-1);
            
            char height_str[20], weight_str[20];
            attron(COLOR_PAIR(3));
            mvprintw(4, 2, "Enter height (cm): ");  // Changed to cm for more intuitive input
            attroff(COLOR_PAIR(3));
            getnstr(height_str, sizeof(height_str)-1);
            height = atof(height_str) / 100.0;  // Convert cm to meters
            
            attron(COLOR_PAIR(3));
            mvprintw(6, 2, "Enter weight (kg): ");
            attroff(COLOR_PAIR(3));
            getnstr(weight_str, sizeof(weight_str)-1);
            weight = atof(weight_str);

            // Input validation
            if (height <= 0 || weight <= 0) {
                attron(COLOR_PAIR(4) | A_BOLD);
                mvprintw(8, 2, "Error: Height and weight must be positive numbers!");
                attroff(COLOR_PAIR(4) | A_BOLD);
                mvprintw(10, 2, "Press any key to try again...");
                refresh();
                getch();
                continue;
            }
            
            if (height > 3) {  // If someone enters height in cm by mistake
                height = height / 100.0;
            }
            
            noecho();

            double bmi = calculate_bmi(weight, height);
            insert_record(db, name, bmi);

            attron(COLOR_PAIR(1) | A_BOLD);
            mvprintw(8, 2, "Results for %s:", name);
            attroff(COLOR_PAIR(1) | A_BOLD);
            
            attron(COLOR_PAIR(2));
            mvprintw(10, 2, "BMI: %.2f", bmi);
            attroff(COLOR_PAIR(2));
            
            attron(COLOR_PAIR(4));
            mvprintw(12, 2, "Advice: %s", bmi_advice(bmi));
            attroff(COLOR_PAIR(4));
            
            attron(COLOR_PAIR(3));
            mvprintw(14, 2, "Press any key to continue...");
            attroff(COLOR_PAIR(3));
            refresh();
            getch();
        }
        else if (choice == 2) {
            view_history(db);
        }
        else if (choice == 3) {
            break;
        }
    }

    endwin();
    sqlite3_close(db);
    return 0;
}
