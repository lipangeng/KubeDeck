package api

import (
	"net/http"
	"sort"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubedeck/backend/internal/k8s"
)

// EventsHandler handles events-related requests
func (h *KernelHandler) EventsHandler(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	if namespace == "" {
		namespace = "default"
	}
	
	limit := 100
	resource := r.URL.Query().Get("resource")
	name := r.URL.Query().Get("name")

	clientManager := k8s.GlobalClientManager()
	clientset, err := clientManager.GetCurrentClient()
	if err != nil {
		writeJSON(w, map[string]interface{}{
			"events": []interface{}{},
			"error": "No cluster configured",
		})
		return
	}

	var events *corev1.EventList
	if resource != "" && name != "" {
		// Get events for specific resource
		events, err = clientset.CoreV1().Events(namespace).List(r.Context(), metav1.ListOptions{
			FieldSelector: "involvedObject.name=" + name + ",involvedObject.kind=" + resource,
		})
	} else {
		// Get all events in namespace
		events, err = clientset.CoreV1().Events(namespace).List(r.Context(), metav1.ListOptions{})
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert to safe format
	type EventInfo struct {
		Name          string `json:"name"`
		Namespace     string `json:"namespace"`
		Type          string `json:"type"`
		Reason        string `json:"reason"`
		Message       string `json:"message"`
		Count         int32  `json:"count"`
		FirstSeen     string `json:"first_seen"`
		LastSeen      string `json:"last_seen"`
		InvolvedObject string `json:"involved_object"`
	}

	eventList := make([]EventInfo, 0, len(events.Items))
	for _, event := range events.Items {
		eventList = append(eventList, EventInfo{
			Name:          event.Name,
			Namespace:     event.Namespace,
			Type:          event.Type,
			Reason:        event.Reason,
			Message:       event.Message,
			Count:         event.Count,
			FirstSeen:     event.FirstTimestamp.Time.Format(time.RFC3339),
			LastSeen:      event.LastTimestamp.Time.Format(time.RFC3339),
			InvolvedObject: event.InvolvedObject.Kind + "/" + event.InvolvedObject.Name,
		})
	}

	// Sort by last seen (newest first)
	sort.Slice(eventList, func(i, j int) bool {
		return eventList[i].LastSeen > eventList[j].LastSeen
	})

	// Apply limit
	if len(eventList) > limit {
		eventList = eventList[:limit]
	}

	writeJSON(w, map[string]interface{}{
		"events": eventList,
		"total":  len(events.Items),
	})
}

// GetEventDetailsHandler gets details for a specific event
func (h *KernelHandler) GetEventDetailsHandler(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	if namespace == "" {
		namespace = "default"
	}
	name := r.URL.Query().Get("name")

	if name == "" {
		http.Error(w, "name required", http.StatusBadRequest)
		return
	}

	clientManager := k8s.GlobalClientManager()
	clientset, err := clientManager.GetCurrentClient()
	if err != nil {
		http.Error(w, "No cluster configured", http.StatusInternalServerError)
		return
	}

	event, err := clientset.CoreV1().Events(namespace).Get(r.Context(), name, metav1.GetOptions{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	type EventDetail struct {
		Name              string            `json:"name"`
		Namespace         string            `json:"namespace"`
		Type              string            `json:"type"`
		Reason            string            `json:"reason"`
		Message           string            `json:"message"`
		Count             int32             `json:"count"`
		FirstSeen         string            `json:"first_seen"`
		LastSeen          string            `json:"last_seen"`
		InvolvedObject    map[string]string `json:"involved_object"`
		Source            map[string]string `json:"source"`
		Annotations       map[string]string `json:"annotations,omitempty"`
	}

	detail := EventDetail{
		Name:      event.Name,
		Namespace: event.Namespace,
		Type:      event.Type,
		Reason:    event.Reason,
		Message:   event.Message,
		Count:     event.Count,
		FirstSeen: event.FirstTimestamp.Time.Format(time.RFC3339),
		LastSeen:  event.LastTimestamp.Time.Format(time.RFC3339),
		InvolvedObject: map[string]string{
			"kind":      event.InvolvedObject.Kind,
			"name":      event.InvolvedObject.Name,
			"apiVersion": event.InvolvedObject.APIVersion,
			"uid":       string(event.InvolvedObject.UID),
		},
		Source: map[string]string{
			"component": event.Source.Component,
			"host":      event.Source.Host,
		},
		Annotations: event.Annotations,
	}

	writeJSON(w, detail)
}
