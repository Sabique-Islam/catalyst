#include <stdio.h>
#include <omp.h>

int main() {
    printf("Testing native Windows OpenMP support\n");
    
    #pragma omp parallel
    {
        int thread_id = omp_get_thread_num();
        int num_threads = omp_get_num_threads();
        
        #pragma omp critical
        {
            printf("Hello from thread %d of %d\n", thread_id, num_threads);
        }
    }
    
    printf("OpenMP test completed successfully!\n");
    return 0;
}