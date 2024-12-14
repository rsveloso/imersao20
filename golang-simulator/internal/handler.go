package internal

import (
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

type RouteCreatedEvent struct {
	EventName  string       `json:"event"`
	RouteID    string       `json:"id"`
	Distance   int          `json:"distance"`
	Directions []Directions `json:"directions"`
}

func NewRouteCreatedEvent(routeID string, distance int, directions []Directions) *RouteCreatedEvent {
	return &RouteCreatedEvent{
		EventName:  "RouteCreated",
		RouteID:    routeID,
		Distance:   distance,
		Directions: directions,
	}
}

type FreightCalculatedEvent struct {
	EventName string  `json:"event"`
	RouteID   string  `json:"route_id"`
	Amount    float64 `json:"amount"`
}

func NewFreightCalculatedEvent(routeID string, amount float64) *FreightCalculatedEvent {
	return &FreightCalculatedEvent{
		EventName: "FreightCalculated",
		RouteID:   routeID,
		Amount:    amount,
	}
}

type DeliveryStartedEvent struct {
	EventName string `json:"event"`
	RouteID   string `json:"route_id"`
}

func NewDeliveryStartedEvent(routeID string) *DeliveryStartedEvent {
	return &DeliveryStartedEvent{
		EventName: "DeliveryStarted",
		RouteID:   routeID,
	}
}

type DriverMovedEvent struct {
	EventName string  `json:"event"`
	RouteID   string  `json:"route_id"`
	Lat       float64 `json:"lat"`
	Lng       float64 `json:"lng"`
}

func NewDriverMovedEvent(routeID string, lat float64, lng float64) *DriverMovedEvent {
	return &DriverMovedEvent{
		EventName: "DriverMoved",
		RouteID:   routeID,
		Lat:       lat,
		Lng:       lng,
	}
}

func RouteCreatedHandler(event *RouteCreatedEvent, routeService *RouteService, mongoClient *mongo.Client) (*FreightCalculatedEvent, error) {
	route := NewRoute(event.RouteID, event.Distance, event.Directions)
	routeCreated, err := routeService.CreateRoute(route)
	if err != nil {
		return nil, err
	}
	freightCalculatedEvent := NewFreightCalculatedEvent(routeCreated.ID, routeCreated.FreightPrice)
	return freightCalculatedEvent, nil
}

func DeliveryStartedHandler(event *DeliveryStartedEvent, routeService *RouteService, mongoClient *mongo.Client, ch chan *DriverMovedEvent) error {
	route, err := routeService.GetRoute(event.RouteID)
	if err != nil {
		return err
	}

	driverMovedEvent := NewDriverMovedEvent(route.ID, 0, 0)
	for _, direction := range route.Directions {
		driverMovedEvent.RouteID = route.ID
		driverMovedEvent.Lat = direction.Lat
		driverMovedEvent.Lng = direction.Lng
		ch <- driverMovedEvent
		time.Sleep(1 * time.Second)
	}
	return nil
}

// event hub to handle events
type EventHub struct {
	routeService        *RouteService
	mongoClient         *mongo.Client
	chDriverMoved       chan *DriverMovedEvent
	chFrieghtCalculated chan *FreightCalculatedEvent
}

func NewEventHub(routeService *RouteService, mongoClient *mongo.Client, chDriverMoved chan *DriverMovedEvent, chFreightCalculated chan *FreightCalculatedEvent) *EventHub {
	return &EventHub{
		routeService:        routeService,
		mongoClient:         mongoClient,
		chDriverMoved:       chDriverMoved,
		chFrieghtCalculated: chFreightCalculated,
	}
}

func (eh *EventHub) HandleEvent(event interface{}) error {
	switch e := event.(type) {
	case RouteCreatedEvent:
		freightCalculatedEvent, err := RouteCreatedHandler(&e, eh.routeService, eh.mongoClient)
		if err != nil {
			return err
		}
		// send the event to the channel
		eh.chFrieghtCalculated <- freightCalculatedEvent
	case DeliveryStartedEvent:
		err := DeliveryStartedHandler(&e, eh.routeService, eh.mongoClient, eh.chDriverMoved)
		if err != nil {
			return err
		}
	}
	return nil
}
