#include <thread>
#include <atomic>
#include <vector>
#include <iostream>

std::atomic<bool> lock;
int number = 0;

void benchTAS() {
    for (int i = 0; i < 200000; i++)
    {
        // lock
        for (;;) {
            //while (lock.load(std::memory_order_acquire));
            bool expected = false;
            if (lock.compare_exchange_weak(expected, true))
                break;
        }

        // work
        number += 1;

        // unlock
        lock.store(false);
    }
}

int main() {
    const int N = 64;
    std::vector<std::thread*> pool;

    for (int i = 0; i < N; i++)
        pool.push_back(new std::thread(benchTAS));

    for (int i = 0; i < N; i++)
        pool[i]->join();

    std::cout << "Result: " << number << std::endl;
}
