import random
import threading
import time

# ==========================================================
# Parâmetros da simulação
# ==========================================================
NUM_PHILOSOPHERS = 5      # N filósofos e N garfos
MEALS_PER_PHILOSOPHER = 3  # quantas vezes cada filósofo deve comer


class Fork:
    """Um garfo compartilhado. Cada garfo é protegido por seu próprio
    Lock: apenas um filósofo pode segurá-lo por vez."""

    def __init__(self, fork_id: int):
        self.mu = threading.Lock()
        self.id = fork_id


def think(philosopher_id: int) -> None:
    print(f"Filósofo {philosopher_id} está PENSANDO...")
    time.sleep((random.randint(0, 79) + 40) / 1000)


def eat(philosopher_id: int) -> None:
    time.sleep((random.randint(0, 79) + 40) / 1000)


def philosopher(philosopher_id: int, left_fork: Fork, right_fork: Fork,
                 seating: threading.Semaphore) -> None:
    """Representa o ciclo de vida de UM filósofo:
    pensar -> tentar pegar os garfos -> comer -> devolver os garfos -> repetir."""

    for meal in range(1, MEALS_PER_PHILOSOPHER + 1):
        think(philosopher_id)

        # (1) Pede permissão ao "garçom" (arbitrador de Dijkstra) antes de
        # disputar os garfos. O semáforo `seating` tem capacidade N-1, ou
        # seja, no máximo N-1 filósofos podem estar tentando pegar garfos ao
        # mesmo tempo. Isso torna IMPOSSÍVEL que os N filósofos fiquem, cada
        # um, com um garfo na mão esperando pelo vizinho — a condição
        # necessária para deadlock (espera circular) nunca se forma.
        seating.acquire()

        # (2) Aquisição ORDENADA dos garfos: sempre trava primeiro o garfo
        # de menor id. Essa é uma segunda barreira independente contra
        # deadlock (quebra de espera circular por ordenação total de
        # recursos), mantida mesmo que a capacidade do "garçom" mude no
        # futuro.
        first, second = left_fork, right_fork
        if first.id > second.id:
            first, second = second, first

        first.mu.acquire()
        second.mu.acquire()

        print(f"Filósofo {philosopher_id} pegou os garfos {first.id} e {second.id} "
              f"e está COMENDO (refeição {meal})")
        eat(philosopher_id)

        second.mu.release()
        first.mu.release()

        # (3) Libera o assento para que outro filósofo possa tentar comer.
        seating.release()

        print(f"Filósofo {philosopher_id} devolveu os garfos {left_fork.id} e "
              f"{right_fork.id} (refeição {meal} concluída)")

    print(f">>> Filósofo {philosopher_id} terminou todas as suas refeições e saiu da mesa.")


def main() -> None:
    random.seed(time.time_ns())

    # Cada garfo é um recurso compartilhado independente, identificado por
    # um índice de 0 a N-1.
    forks = [Fork(i) for i in range(NUM_PHILOSOPHERS)]

    # Semáforo contador usado como mecanismo de sincronização: representa o
    # "garçom" do problema clássico. Capacidade N-1 garante que ao menos um
    # filósofo sempre consiga liberar um garfo, prevenindo deadlock.
    seating = threading.Semaphore(NUM_PHILOSOPHERS - 1)

    # Cada filósofo é uma thread independente, concorrendo pelos dois garfos
    # vizinhos (esquerdo e direito) dispostos em círculo.
    threads = []
    for i in range(NUM_PHILOSOPHERS):
        left = forks[i]
        right = forks[(i + 1) % NUM_PHILOSOPHERS]
        t = threading.Thread(target=philosopher, args=(i, left, right, seating))
        threads.append(t)
        t.start()

    # A thread principal aguarda todas as threads dos filósofos terminarem
    # antes de encerrar o programa (equivalente ao wg.Wait() do Go).
    for t in threads:
        t.join()

    print("Todos os filósofos terminaram. Simulação encerrada com sucesso.")


if __name__ == "__main__":
    main()
