# MfNF-Collaborator-Numbers
Returns an sql insert/update-statement giving the number edits of user "xyz" on day "asd" in topic "qwe" to generate an overview of all collaborators and their edits for the Wikibooks-project ["Mathe für Nicht-Freaks"](https://de.wikibooks.org/wiki/Mathe_f%C3%BCr_Nicht-Freaks). This is based on the [Sitemap](https://de.wikibooks.org/w/index.php?title=Mathe_f%C3%BCr_Nicht-Freaks:_Sitemap)

Assumed database-structure

```SQL
CREATE TABLE MFNF_EDITS (
	id INT(11) NOT NULL AUTO_INCREMENT,
	date DATE,
	name CHAR(255),
	topic CHAR(255),
	number_of_edits INT(11),
	PRIMARY KEY ( id ),
	UNIQUE (date, name, topic)
)
```

Topics to be retrieved (in config.json): 

```json
{
    "topics": ["Grundlagen der Mathematik", "Analysis 1", "Lineare Algebra 1","Maßtheorie","Real Analysis", "Mitmachen für (Nicht-)Freaks"]
}
 
```
