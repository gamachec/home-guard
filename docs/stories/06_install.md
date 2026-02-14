### Story 6 : Persistance et Auto-lancement (Finalisation)

**Objectif :** Rendre l'agent résilient et invisible.

* **Détails techniques :**
    * Ajouter une commande permettant l'installation (ajouter l'exécutable en tant que service Windows)
    * S'assurer que l'agent gère les pertes de connexion Wi-Fi/MQTT en tentant des reconnexions automatiques (
      exponential backoff).