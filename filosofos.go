package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// ==========================================================
// Parâmetros da simulação
// ==========================================================
const (
	numPhilosophers     = 5 // N filósofos e N garfos
	mealsPerPhilosopher = 3 // quantas vezes cada filósofo deve comer
)

// Fork representa um garfo compartilhado. Cada garfo é protegido por seu
// próprio Mutex: apenas um filósofo pode segurá-lo por vez.
type Fork struct {
	mu sync.Mutex
	id int
}

// philosopher é a goroutine que representa o ciclo de vida de UM filósofo:
// pensar -> tentar pegar os garfos -> comer -> devolver os garfos -> repetir.
func philosopher(id int, leftFork, rightFork *Fork, seating chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	for meal := 1; meal <= mealsPerPhilosopher; meal++ {
		think(id)

		// (1) Pede permissão ao "garçom" (arbitrador de Dijkstra) antes de
		// disputar os garfos. O canal `seating` tem capacidade N-1, ou seja,
		// no máximo N-1 filósofos podem estar tentando pegar garfos ao mesmo
		// tempo. Isso torna IMPOSSÍVEL que os N filósofos fiquem, cada um,
		// com um garfo na mão esperando pelo vizinho — a condição necessária
		// para deadlock (espera circular) nunca se forma.
		seating <- struct{}{}

		// (2) Aquisição ORDENADA dos garfos: sempre trava primeiro o garfo
		// de menor id. Essa é uma segunda barreira independente contra
		// deadlock (quebra de espera circular por ordenação total de
		// recursos), mantida mesmo que a capacidade do "garçom" mude no
		// futuro.
		first, second := leftFork, rightFork
		if first.id > second.id {
			first, second = second, first
		}

		first.mu.Lock()
		second.mu.Lock()

		fmt.Printf("Filósofo %d pegou os garfos %d e %d e está COMENDO (refeição %d)\n",
			id, first.id, second.id, meal)
		eat(id)

		second.mu.Unlock()
		first.mu.Unlock()

		// (3) Libera o assento para que outro filósofo possa tentar comer.
		<-seating

		fmt.Printf("Filósofo %d devolveu os garfos %d e %d (refeição %d concluída)\n",
			id, leftFork.id, rightFork.id, meal)
	}

	fmt.Printf(">>> Filósofo %d terminou todas as suas refeições e saiu da mesa.\n", id)
}

func think(id int) {
	fmt.Printf("Filósofo %d está PENSANDO...\n", id)
	time.Sleep(time.Duration(rand.Intn(80)+40) * time.Millisecond)
}

func eat(id int) {
	time.Sleep(time.Duration(rand.Intn(80)+40) * time.Millisecond)
}

func main() {
	rand.Seed(time.Now().UnixNano())

	// Cada garfo é um recurso compartilhado independente, identificado por
	// um índice de 0 a N-1.
	forks := make([]*Fork, numPhilosophers)
	for i := 0; i < numPhilosophers; i++ {
		forks[i] = &Fork{id: i}
	}

	// Canal usado como semáforo contador (não para trocar dados, e sim como
	// mecanismo de sincronização): representa o "garçom" do problema
	// clássico. Capacidade N-1 garante que ao menos um filósofo sempre
	// consiga liberar um garfo, prevenindo deadlock.
	seating := make(chan struct{}, numPhilosophers-1)

	var wg sync.WaitGroup
	wg.Add(numPhilosophers)

	// Cada filósofo é uma goroutine independente, concorrendo pelos dois
	// garfos vizinhos (esquerdo e direito) dispostos em círculo.
	for i := 0; i < numPhilosophers; i++ {
		left := forks[i]
		right := forks[(i+1)%numPhilosophers]
		go philosopher(i, left, right, seating, &wg)
	}

	// A goroutine principal aguarda todas as goroutines dos filósofos
	// terminarem antes de encerrar o programa.
	wg.Wait()
	fmt.Println("Todos os filósofos terminaram. Simulação encerrada com sucesso.")
}
