### Story 3 : Système de Notifications Windows Natif

**Objectif :** Permettre à l'agent d'afficher des messages à l'utilisateur sans interface graphique propre.

* **Détails techniques :**
    * Utiliser une bibliothèque Go pour les "Windows Toast Notifications" (ex: `github.com/go-toast/toast`).
    * S'assurer que l'agent peut afficher un titre, un message et éventuellement une icône.
    * **Test :** Créer un topic MQTT `cmnd/agent_pc/notify` qui affiche le contenu du message reçu sur l'écran de l'
      enfant.
