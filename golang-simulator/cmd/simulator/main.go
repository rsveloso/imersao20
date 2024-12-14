package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/devfullcycle/imersao20/simulator/internal"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	mongoStr := "mongodb://admin:admin@localhost:27017/routes?authSource=admin"
	// open new connection with mongodb
	mongoConnection, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoStr))
	if err != nil {
		panic(err)
	}

	/// create a new route service
	freightService := internal.NewFreightService()
	routeService := internal.NewRouteService(mongoConnection, freightService)
	chDriverMoved := make(chan *internal.DriverMovedEvent)
	chFreightCalculated := make(chan *internal.FreightCalculatedEvent)

	hub := internal.NewEventHub(
		routeService,
		mongoConnection,
		chDriverMoved,
		chFreightCalculated,
	)

	// create some events to test
	routeCreatedEvent := internal.RouteCreatedEvent{
		RouteID:  "1",
		Distance: 100,
		Directions: []internal.Directions{
			{Lat: 1.0, Lng: 1.0},
			{Lat: 2.0, Lng: 2.0},
			{Lat: 3.0, Lng: 3.0},
		},
	}

	deliveryStartedEvent := internal.DeliveryStartedEvent{
		RouteID: "1",
	}

	// handle the events
	go hub.HandleEvent(routeCreatedEvent)
	go hub.HandleEvent(deliveryStartedEvent)

	// receive the events from channels needs to process all events
	// to avoid deadlock
	for {
		select {
		case driverMovedEvent := <-chDriverMoved:
			jsonDriverMoved, _ := json.Marshal(driverMovedEvent)
			fmt.Println(string(jsonDriverMoved))
		case freightCalculatedEvent := <-chFreightCalculated:
			jsonFreightCalculated, _ := json.Marshal(freightCalculatedEvent)
			fmt.Println(string(jsonFreightCalculated))
		}
	}
}
