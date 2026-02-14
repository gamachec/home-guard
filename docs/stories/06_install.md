### Story 6 : Persistance et Auto-lancement (Finalisation)

**Objectif :** Rendre l'agent résilient et invisible.

* **Détails techniques :**
    * Ajouter une commande (ou un script d'installation) pour inscrire l'exécutable dans la clé de registre
      `HKEY_CURRENT_USER\Software\Microsoft\Windows\CurrentVersion\Run`.
    * S'assurer que l'agent gère les pertes de connexion Wi-Fi/MQTT en tentant des reconnexions automatiques (
      exponential backoff).
    * Optimiser la boucle de scan (CPU usage) en utilisant des timers (`time.Ticker`) plutôt que des pauses `time.Sleep`
      bloquantes.
