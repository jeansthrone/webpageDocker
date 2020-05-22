package main

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"regexp"
)

//Los registros representan páginas
type Pagina struct {
	Titulo string
	Cuerpo []byte
}

type Usersi struct {
	Name     string
	Lastname string
}

var plantillas = template.Must(template.ParseFiles("tmpl/edit.html", "tmpl/view.html", "tmpl/front.html", "tmpl/insertar.html"))
var regex_ruta = regexp.MustCompile("^(/|(/(edit|save|view|election|saveu|insertaru)/([a-zA-Z0-9]+)))?$")
var pagina_principal = "Principal2"

func main() {
	//Creamos y guardamos una página en disco.
	pag1 := &Pagina{Titulo: "Ejemplo", Cuerpo: []byte("Este es el cuerpo")}
	pag1.guardar()

	http.HandleFunc("/", llamarManejador(manejadorRaiz)) //Nuevo manejador
	http.HandleFunc("/view/", llamarManejador(manejadorMostrar))
	http.HandleFunc("/save/", llamarManejador(manejadorGuardar))
	http.HandleFunc("/edit/", llamarManejador(manejadorEditar))
	http.HandleFunc("/election/", llamarManejador(manejadorElection))
	http.HandleFunc("/saveu/", llamarManejador(manejadorGuardaru))
	http.HandleFunc("/insertaru/", llamarManejador(manejadorInsertaru))
	fmt.Println("El servidor se encuentra en ejecución")
	http.ListenAndServe(":8080", nil)
}

//Este método almacenará páginas en disco duro
func (p *Pagina) guardar() error {
	nombre := p.Titulo + ".txt"
	return ioutil.WriteFile("./view/"+nombre, p.Cuerpo, 0600)
}

//Leer páginas
func cargarPagina(titulo string) (*Pagina, error) {
	nombre_archivo := titulo + ".txt"
	fmt.Println("El cliente ha pedido:" + nombre_archivo)
	cuerpo, err := ioutil.ReadFile("./view/" + nombre_archivo)

	if err != nil {
		return nil, err
	}
	return &Pagina{Titulo: titulo, Cuerpo: cuerpo}, nil

}

func obtenerBaseDeDatos() (db *sql.DB, e error) {
	usuario := "docker"
	pass := "docker"
	host := "tcp(192.168.99.100:3306)"
	nombreBaseDeDatos := "prueba"
	// Debe tener la forma usuario:contraseña@host/nombreBaseDeDatos
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@%s/%s", usuario, pass, host, nombreBaseDeDatos))
	if err != nil {
		return nil, err
	}
	return db, nil
}

func GetUsuarios() ([]Usersi, error) {
	users := []Usersi{}
	db, err := obtenerBaseDeDatos()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	filas, err := db.Query("SELECT name, lastname FROM users")
	if err != nil {
		return nil, err
	}
	defer filas.Close()

	var c Usersi

	for filas.Next() {
		err2 := filas.Scan(&c.Name, &c.Lastname)
		if err2 != nil {
			return nil, err2
		}

		users = append(users, c)

	}

	return users, nil
}

//Función para validar ruta y regresar nombre de la página solicitada
func dameTitulo(w http.ResponseWriter, r *http.Request) (string, error) {
	peticion := regex_ruta.FindStringSubmatch(r.URL.Path)
	if peticion == nil {
		http.NotFound(w, r)
		return "", errors.New("Ruta inválida")
	}
	return peticion[len(peticion)-1], nil
}

//Carga las plantillas HTML
func cargarPlantilla(w http.ResponseWriter, nombre_plantilla string, p *Pagina) {
	plantillas.ExecuteTemplate(w, nombre_plantilla+".html", p)
}
func cargarPlantillaMostrar(w http.ResponseWriter, nombre_plantilla string, p *Pagina, u []Usersi) {

	users := u
	fmt.Println(users)
	i := map[string]interface{}{
		"Titulo": p.Titulo,
		"User":   users,
	}

	tmpls, _ := template.ParseFiles(filepath.Join("tmpl", "view.html"))
	tmpl := tmpls.Lookup("view.html")
	tmpl.Execute(w, i)
	/*fmt.Println(i)
	plantillas.ExecuteTemplate(w, nombre_plantilla+".html", i)*/
}

//Closure para que redireccione a donde debe sin importar si colocamos mal la ruta
func llamarManejador(manejador func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		titulo, err := dameTitulo(w, r)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		manejador(w, r, titulo)
	}
}

//Manejador para mostrar página principal
func manejadorRaiz(w http.ResponseWriter, r *http.Request, titulo string) {
	p, err := cargarPagina(pagina_principal)
	if err != nil {
		http.Redirect(w, r, "edit/"+pagina_principal, http.StatusFound)
		fmt.Println("Error")
		return
	}
	cargarPlantilla(w, "front", p)
}

//Manejador de peticiones
func manejadorMostrar(w http.ResponseWriter, r *http.Request, titulo string) {
	p, err := cargarPagina(titulo)
	u, err := GetUsuarios()
	if err != nil {
		http.Redirect(w, r, "/edit/"+titulo, http.StatusFound)
		fmt.Println("La página solicitada no existía. Llamando al editor...")
		return
	}
	cargarPlantillaMostrar(w, "view", p, u)
}

//Manejador para editar wikis
func manejadorEditar(w http.ResponseWriter, r *http.Request, titulo string) {
	p, err := cargarPagina(titulo)
	if err != nil {
		p = &Pagina{Titulo: titulo}
	}
	cargarPlantilla(w, "edit", p)
}

//Manejador para guardar wikis
func manejadorGuardar(w http.ResponseWriter, r *http.Request, titulo string) {
	cuerpo := r.FormValue("body")
	p := &Pagina{Titulo: titulo, Cuerpo: []byte(cuerpo)}
	fmt.Println("Guardando " + titulo + ".txt...")
	p.guardar()
	http.Redirect(w, r, "/view/"+titulo, http.StatusFound)
}

/*********************************** New *****************************************/

type Users struct {
	Name []byte
}

func (p *Users) guardaru() error {

	return ioutil.WriteFile("./view/Users.txt", p.Name, 0600)
}

func (p *Usersi) insertaru() error {
	db, err := obtenerBaseDeDatos()
	if err != nil {
		fmt.Printf("Error obteniendo base de datos: %v", err)
		return err
	}
	defer db.Close()
	// Preparamos para prevenir inyecciones SQL
	sentenciaPreparada, err := db.Prepare("INSERT INTO users(name, lastname) VALUES( ?, ? )")
	if err != nil {
		return err
	}
	defer sentenciaPreparada.Close()
	// Ejecutar sentencia, un valor por cada '?'
	_, err = sentenciaPreparada.Exec(p.Name, p.Lastname)
	if err != nil {
		return err
	}
	fmt.Println("Guardado exitoso!")
	return nil

}

func manejadorElection(w http.ResponseWriter, r *http.Request, titulo string) {
	keys, ok := r.URL.Query()["value"]
	if !ok || len(keys[0]) < 1 {
		log.Println("Url Param 'key' is missing")
		return
	}
	key := keys[0]
	/*cuerpo := r.FormValue("body")*/
	fmt.Println(key)
	p := &Pagina{Titulo: titulo}

	switch key {
	case "1":
		cargarPlantilla(w, "insertar", p)
	case "2":
		http.Redirect(w, r, "/view/Users", http.StatusFound)
	case "3":
		http.Redirect(w, r, "/edit/Users", http.StatusFound)
	}

}

func manejadorInsertaru(w http.ResponseWriter, r *http.Request, titulo string) {
	name := r.FormValue("name")
	last := r.FormValue("last")

	p := &Usersi{Name: name, Lastname: last}
	fmt.Println("Guardando " + titulo + ".txt...")
	p.insertaru()
	http.Redirect(w, r, "/", http.StatusFound)
}

func manejadorGuardaru(w http.ResponseWriter, r *http.Request, titulo string) {
	name := r.FormValue("name")
	last := r.FormValue("last")
	by := []byte(name + " " + last)

	p := &Users{Name: by}
	fmt.Println("Guardando " + titulo + ".txt...")
	p.guardaru()
	http.Redirect(w, r, "/", http.StatusFound)
}
