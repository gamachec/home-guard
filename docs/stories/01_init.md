### Story 1 : Fondations et Connectivité MQTT

**Objectif :** Créer un agent Go qui tourne en arrière-plan, se connecte à ton broker MQTT et annonce sa présence.

* **Détails techniques :**
    * Initialiser un projet Go avec le client `paho.mqtt.golang`.
    * Implémenter une structure de configuration (JSON ou YAML) pour l'adresse du broker, les identifiants et le
      `client_id` (nom du PC).
    * Mettre en place le "LWT" (Last Will and Testament) sur le topic `stat/agent_pc/status` ("online"/"offline").
    * L'agent doit tourner dans une boucle infinie (ou via un signal système) pour ne pas s'arrêter.
    * Compiler en mode "windowsgui" pour que la console n'apparaisse pas au lancement.