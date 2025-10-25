#include <stdio.h>
#include <omp.h>

int main() {
    int n = 1000000;
    double sum = 0.0;
    double start_time, end_time;

    // Create an array of numbers
    double arr[n];
    for (int i = 0; i < n; i++) {
        arr[i] = i * 0.5;
    }

    start_time = omp_get_wtime(); // Start timing

    // Parallel reduction: sum the array
    #pragma omp parallel for reduction(+:sum)
    for (int i = 0; i < n; i++) {
        sum += arr[i];
    }

    end_time = omp_get_wtime(); // End timing

    printf("Sum = %.2f\n", sum);
    printf("Time taken = %.6f seconds\n", end_time - start_time);
    printf("Threads used = %d\n", omp_get_max_threads());

    return 0;
}
