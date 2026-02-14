### Story 5 : Gestion de la "Blacklist" dynamique

**Objectif :** Pouvoir modifier la liste des applications interdites sans recompiler l'agent.

* **Détails techniques :**
    * Gérer une liste de chaînes de caractères en mémoire.
    * Écouter sur le topic `cmnd/agent_pc/blacklist/set` (format JSON : `["game1.exe", "browser.exe"]`).
    * Sauvegarder cette liste dans un fichier local (ex: `config.json`) pour que l'agent s'en souvienne après un
      redémarrage.
    * S'assurer que la boucle de "Kill" de la Story 4 utilise bien cette liste mise à jour.
