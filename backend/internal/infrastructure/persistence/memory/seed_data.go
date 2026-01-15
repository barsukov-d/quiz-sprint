package memory

import (
	"time"

	"github.com/barsukov/quiz-sprint/backend/internal/domain/quiz"
)

// seedData populates the repository with test quizzes
func (r *QuizRepository) seedData() {
	now := time.Now().Unix()

	// Quiz 1: JavaScript Basics
	r.createJavaScriptQuiz(now)

	// Quiz 2: World Geography
	r.createGeographyQuiz(now)

	// Quiz 3: Movie Trivia
	r.createMovieQuiz(now)

	// Quiz 4: General Knowledge
	r.createGeneralKnowledgeQuiz(now)
}

// createJavaScriptQuiz creates a JavaScript programming quiz
func (r *QuizRepository) createJavaScriptQuiz(now int64) {
	quizID := quiz.NewQuizID()
	title, _ := quiz.NewQuizTitle("JavaScript Basics")
	timeLimit, _ := quiz.NewTimeLimit(300) // 5 minutes
	passingScore, _ := quiz.NewPassingScore(70)

	jsQuiz, _ := quiz.NewQuiz(
		quizID,
		title,
		"Test your fundamental knowledge of JavaScript including ES6+ features, async/await, and closures",
		timeLimit,
		passingScore,
		now,
	)

	// Question 1
	q1, _ := createQuestion(
		"What does the 'const' keyword do in JavaScript?",
		10,
		1,
		[]answerData{
			{"Creates a constant variable that cannot be reassigned", true},
			{"Creates a variable that can be reassigned", false},
			{"Creates a global variable", false},
			{"Creates a private variable", false},
		},
	)
	jsQuiz.AddQuestion(*q1)

	// Question 2
	q2, _ := createQuestion(
		"What is a closure in JavaScript?",
		10,
		2,
		[]answerData{
			{"A function with access to its outer scope even after the outer function has returned", true},
			{"An async function", false},
			{"A loop structure", false},
			{"A class method", false},
		},
	)
	jsQuiz.AddQuestion(*q2)

	// Question 3
	q3, _ := createQuestion(
		"What is the purpose of 'async/await'?",
		10,
		3,
		[]answerData{
			{"To handle asynchronous operations in a synchronous-looking way", true},
			{"To create classes", false},
			{"To define constants", false},
			{"To create loops", false},
		},
	)
	jsQuiz.AddQuestion(*q3)

	// Question 4
	q4, _ := createQuestion(
		"What does the spread operator (...) do?",
		10,
		4,
		[]answerData{
			{"Expands an iterable into individual elements", true},
			{"Creates a new variable", false},
			{"Defines a function", false},
			{"Creates a loop", false},
		},
	)
	jsQuiz.AddQuestion(*q4)

	// Question 5
	q5, _ := createQuestion(
		"What is the difference between '==' and '==='?",
		10,
		5,
		[]answerData{
			{"=== checks both value and type, == only checks value", true},
			{"They are the same", false},
			{"== is faster than ===", false},
			{"=== is deprecated", false},
		},
	)
	jsQuiz.AddQuestion(*q5)

	// Question 6
	q6, _ := createQuestion(
		"What is the purpose of 'map()' method?",
		10,
		6,
		[]answerData{
			{"Creates a new array by transforming each element", true},
			{"Filters array elements", false},
			{"Sorts array elements", false},
			{"Reverses array elements", false},
		},
	)
	jsQuiz.AddQuestion(*q6)

	// Question 7
	q7, _ := createQuestion(
		"What is a Promise in JavaScript?",
		10,
		7,
		[]answerData{
			{"An object representing the eventual completion or failure of an async operation", true},
			{"A synchronous function", false},
			{"A loop construct", false},
			{"A variable type", false},
		},
	)
	jsQuiz.AddQuestion(*q7)

	// Question 8
	q8, _ := createQuestion(
		"What does 'this' keyword refer to?",
		10,
		8,
		[]answerData{
			{"The object that is executing the current function", true},
			{"Always the global object", false},
			{"Always null", false},
			{"Always undefined", false},
		},
	)
	jsQuiz.AddQuestion(*q8)

	// Question 9
	q9, _ := createQuestion(
		"What is destructuring in JavaScript?",
		10,
		9,
		[]answerData{
			{"A way to extract values from arrays or objects into variables", true},
			{"A way to delete variables", false},
			{"A way to create classes", false},
			{"A way to define functions", false},
		},
	)
	jsQuiz.AddQuestion(*q9)

	// Question 10
	q10, _ := createQuestion(
		"What is the event loop?",
		10,
		10,
		[]answerData{
			{"A mechanism that handles asynchronous callbacks", true},
			{"A type of loop structure", false},
			{"A way to define events", false},
			{"A DOM manipulation method", false},
		},
	)
	jsQuiz.AddQuestion(*q10)

	r.quizzes[quizID.String()] = jsQuiz
}

// createGeographyQuiz creates a world geography quiz
func (r *QuizRepository) createGeographyQuiz(now int64) {
	quizID := quiz.NewQuizID()
	title, _ := quiz.NewQuizTitle("World Geography")
	timeLimit, _ := quiz.NewTimeLimit(180) // 3 minutes
	passingScore, _ := quiz.NewPassingScore(60)

	geoQuiz, _ := quiz.NewQuiz(
		quizID,
		title,
		"Test your knowledge of world capitals, countries, and landmarks",
		timeLimit,
		passingScore,
		now,
	)

	// Question 1
	q1, _ := createQuestion(
		"What is the capital of France?",
		10,
		1,
		[]answerData{
			{"Paris", true},
			{"London", false},
			{"Berlin", false},
			{"Madrid", false},
		},
	)
	geoQuiz.AddQuestion(*q1)

	// Question 2
	q2, _ := createQuestion(
		"Which country has the largest population?",
		10,
		2,
		[]answerData{
			{"India", true},
			{"United States", false},
			{"Russia", false},
			{"Brazil", false},
		},
	)
	geoQuiz.AddQuestion(*q2)

	// Question 3
	q3, _ := createQuestion(
		"What is the longest river in the world?",
		10,
		3,
		[]answerData{
			{"Nile River", true},
			{"Amazon River", false},
			{"Yangtze River", false},
			{"Mississippi River", false},
		},
	)
	geoQuiz.AddQuestion(*q3)

	// Question 4
	q4, _ := createQuestion(
		"Which continent is the Sahara Desert located in?",
		10,
		4,
		[]answerData{
			{"Africa", true},
			{"Asia", false},
			{"Australia", false},
			{"South America", false},
		},
	)
	geoQuiz.AddQuestion(*q4)

	// Question 5
	q5, _ := createQuestion(
		"What is the smallest country in the world?",
		10,
		5,
		[]answerData{
			{"Vatican City", true},
			{"Monaco", false},
			{"San Marino", false},
			{"Liechtenstein", false},
		},
	)
	geoQuiz.AddQuestion(*q5)

	// Question 6
	q6, _ := createQuestion(
		"Mount Everest is located in which mountain range?",
		10,
		6,
		[]answerData{
			{"Himalayas", true},
			{"Alps", false},
			{"Andes", false},
			{"Rocky Mountains", false},
		},
	)
	geoQuiz.AddQuestion(*q6)

	// Question 7
	q7, _ := createQuestion(
		"What is the capital of Japan?",
		10,
		7,
		[]answerData{
			{"Tokyo", true},
			{"Seoul", false},
			{"Beijing", false},
			{"Bangkok", false},
		},
	)
	geoQuiz.AddQuestion(*q7)

	// Question 8
	q8, _ := createQuestion(
		"Which ocean is the largest?",
		10,
		8,
		[]answerData{
			{"Pacific Ocean", true},
			{"Atlantic Ocean", false},
			{"Indian Ocean", false},
			{"Arctic Ocean", false},
		},
	)
	geoQuiz.AddQuestion(*q8)

	// Question 9
	q9, _ := createQuestion(
		"The Great Barrier Reef is located off the coast of which country?",
		10,
		9,
		[]answerData{
			{"Australia", true},
			{"Brazil", false},
			{"Indonesia", false},
			{"Philippines", false},
		},
	)
	geoQuiz.AddQuestion(*q9)

	// Question 10
	q10, _ := createQuestion(
		"What is the capital of Canada?",
		10,
		10,
		[]answerData{
			{"Ottawa", true},
			{"Toronto", false},
			{"Vancouver", false},
			{"Montreal", false},
		},
	)
	geoQuiz.AddQuestion(*q10)

	// Questions 11-15 for more challenge
	q11, _ := createQuestion(
		"Which desert is the largest hot desert in the world?",
		10,
		11,
		[]answerData{
			{"Sahara Desert", true},
			{"Arabian Desert", false},
			{"Gobi Desert", false},
			{"Kalahari Desert", false},
		},
	)
	geoQuiz.AddQuestion(*q11)

	q12, _ := createQuestion(
		"What is the currency of the United Kingdom?",
		10,
		12,
		[]answerData{
			{"Pound Sterling", true},
			{"Euro", false},
			{"Dollar", false},
			{"Franc", false},
		},
	)
	geoQuiz.AddQuestion(*q12)

	q13, _ := createQuestion(
		"Which country has the most time zones?",
		10,
		13,
		[]answerData{
			{"France", true},
			{"Russia", false},
			{"United States", false},
			{"China", false},
		},
	)
	geoQuiz.AddQuestion(*q13)

	q14, _ := createQuestion(
		"The Amazon Rainforest is primarily located in which country?",
		10,
		14,
		[]answerData{
			{"Brazil", true},
			{"Colombia", false},
			{"Peru", false},
			{"Venezuela", false},
		},
	)
	geoQuiz.AddQuestion(*q14)

	q15, _ := createQuestion(
		"What is the official language of Brazil?",
		10,
		15,
		[]answerData{
			{"Portuguese", true},
			{"Spanish", false},
			{"English", false},
			{"French", false},
		},
	)
	geoQuiz.AddQuestion(*q15)

	r.quizzes[quizID.String()] = geoQuiz
}

// createMovieQuiz creates a movie trivia quiz
func (r *QuizRepository) createMovieQuiz(now int64) {
	quizID := quiz.NewQuizID()
	title, _ := quiz.NewQuizTitle("Movie Trivia")
	timeLimit, _ := quiz.NewTimeLimit(600) // 10 minutes
	passingScore, _ := quiz.NewPassingScore(65)

	movieQuiz, _ := quiz.NewQuiz(
		quizID,
		title,
		"Test your knowledge of famous films, actors, and cinema history",
		timeLimit,
		passingScore,
		now,
	)

	// Question 1
	q1, _ := createQuestion(
		"Who directed the movie 'The Godfather'?",
		10,
		1,
		[]answerData{
			{"Francis Ford Coppola", true},
			{"Martin Scorsese", false},
			{"Steven Spielberg", false},
			{"Quentin Tarantino", false},
		},
	)
	movieQuiz.AddQuestion(*q1)

	// Question 2
	q2, _ := createQuestion(
		"Which movie won the Oscar for Best Picture in 1994?",
		10,
		2,
		[]answerData{
			{"Forrest Gump", true},
			{"Pulp Fiction", false},
			{"The Shawshank Redemption", false},
			{"The Lion King", false},
		},
	)
	movieQuiz.AddQuestion(*q2)

	// Question 3
	q3, _ := createQuestion(
		"Who played Iron Man in the Marvel Cinematic Universe?",
		10,
		3,
		[]answerData{
			{"Robert Downey Jr.", true},
			{"Chris Evans", false},
			{"Chris Hemsworth", false},
			{"Mark Ruffalo", false},
		},
	)
	movieQuiz.AddQuestion(*q3)

	// Question 4
	q4, _ := createQuestion(
		"What year was the first 'Star Wars' movie released?",
		10,
		4,
		[]answerData{
			{"1977", true},
			{"1980", false},
			{"1975", false},
			{"1983", false},
		},
	)
	movieQuiz.AddQuestion(*q4)

	// Question 5
	q5, _ := createQuestion(
		"Which movie features the line 'Here's looking at you, kid'?",
		10,
		5,
		[]answerData{
			{"Casablanca", true},
			{"Gone with the Wind", false},
			{"The Maltese Falcon", false},
			{"The Big Sleep", false},
		},
	)
	movieQuiz.AddQuestion(*q5)

	// Add more questions for a longer quiz
	questions := []struct {
		text    string
		answers []answerData
	}{
		{
			"Who directed 'Inception'?",
			[]answerData{
				{"Christopher Nolan", true},
				{"Denis Villeneuve", false},
				{"Ridley Scott", false},
				{"James Cameron", false},
			},
		},
		{
			"Which movie won the first ever Oscar for Best Animated Feature?",
			[]answerData{
				{"Shrek", true},
				{"Toy Story", false},
				{"Finding Nemo", false},
				{"Monsters, Inc.", false},
			},
		},
		{
			"Who played the Joker in 'The Dark Knight'?",
			[]answerData{
				{"Heath Ledger", true},
				{"Joaquin Phoenix", false},
				{"Jack Nicholson", false},
				{"Jared Leto", false},
			},
		},
		{
			"What is the highest-grossing film of all time (not adjusted for inflation)?",
			[]answerData{
				{"Avatar", true},
				{"Avengers: Endgame", false},
				{"Titanic", false},
				{"Star Wars: The Force Awakens", false},
			},
		},
		{
			"Which movie features the quote 'May the Force be with you'?",
			[]answerData{
				{"Star Wars", true},
				{"Star Trek", false},
				{"Blade Runner", false},
				{"The Matrix", false},
			},
		},
		{
			"Who directed 'Pulp Fiction'?",
			[]answerData{
				{"Quentin Tarantino", true},
				{"Martin Scorsese", false},
				{"The Coen Brothers", false},
				{"Paul Thomas Anderson", false},
			},
		},
		{
			"Which actress won an Oscar for 'La La Land'?",
			[]answerData{
				{"Emma Stone", true},
				{"Jennifer Lawrence", false},
				{"Natalie Portman", false},
				{"Amy Adams", false},
			},
		},
		{
			"What is the name of the fictional African country in 'Black Panther'?",
			[]answerData{
				{"Wakanda", true},
				{"Zamunda", false},
				{"Genovia", false},
				{"Latveria", false},
			},
		},
		{
			"Which film won the Palme d'Or at Cannes in 2019?",
			[]answerData{
				{"Parasite", true},
				{"Once Upon a Time in Hollywood", false},
				{"Portrait of a Lady on Fire", false},
				{"Pain and Glory", false},
			},
		},
		{
			"Who composed the score for 'The Lord of the Rings' trilogy?",
			[]answerData{
				{"Howard Shore", true},
				{"Hans Zimmer", false},
				{"John Williams", false},
				{"Danny Elfman", false},
			},
		},
		{
			"Which movie features a young Natalie Portman as Mathilda?",
			[]answerData{
				{"Léon: The Professional", true},
				{"The Professional", false},
				{"Heat", false},
				{"Taxi Driver", false},
			},
		},
		{
			"What was the first Pixar movie?",
			[]answerData{
				{"Toy Story", true},
				{"A Bug's Life", false},
				{"Monsters, Inc.", false},
				{"Finding Nemo", false},
			},
		},
		{
			"Who directed 'Schindler's List'?",
			[]answerData{
				{"Steven Spielberg", true},
				{"Roman Polanski", false},
				{"Martin Scorsese", false},
				{"Francis Ford Coppola", false},
			},
		},
		{
			"Which James Bond actor appeared in the most films?",
			[]answerData{
				{"Roger Moore", true},
				{"Sean Connery", false},
				{"Daniel Craig", false},
				{"Pierce Brosnan", false},
			},
		},
		{
			"What is the name of the hotel in 'The Shining'?",
			[]answerData{
				{"The Overlook Hotel", true},
				{"The Stanley Hotel", false},
				{"The Grand Budapest Hotel", false},
				{"The Continental", false},
			},
		},
	}

	for i, q := range questions {
		question, _ := createQuestion(q.text, 10, i+6, q.answers)
		movieQuiz.AddQuestion(*question)
	}

	r.quizzes[quizID.String()] = movieQuiz
}

// createGeneralKnowledgeQuiz creates a general knowledge quiz
func (r *QuizRepository) createGeneralKnowledgeQuiz(now int64) {
	quizID := quiz.NewQuizID()
	title, _ := quiz.NewQuizTitle("General Knowledge Challenge")
	timeLimit, _ := quiz.NewTimeLimit(420) // 7 minutes
	passingScore, _ := quiz.NewPassingScore(70)

	gkQuiz, _ := quiz.NewQuiz(
		quizID,
		title,
		"Test your general knowledge across various topics including history, science, and culture",
		timeLimit,
		passingScore,
		now,
	)

	questions := []struct {
		text    string
		answers []answerData
	}{
		{
			"What is the speed of light?",
			[]answerData{
				{"299,792,458 meters per second", true},
				{"150,000,000 meters per second", false},
				{"500,000,000 meters per second", false},
				{"100,000,000 meters per second", false},
			},
		},
		{
			"Who wrote 'Romeo and Juliet'?",
			[]answerData{
				{"William Shakespeare", true},
				{"Charles Dickens", false},
				{"Jane Austen", false},
				{"Mark Twain", false},
			},
		},
		{
			"What is the chemical symbol for gold?",
			[]answerData{
				{"Au", true},
				{"Ag", false},
				{"Fe", false},
				{"Cu", false},
			},
		},
		{
			"In what year did World War II end?",
			[]answerData{
				{"1945", true},
				{"1944", false},
				{"1946", false},
				{"1943", false},
			},
		},
		{
			"What is the largest planet in our solar system?",
			[]answerData{
				{"Jupiter", true},
				{"Saturn", false},
				{"Neptune", false},
				{"Uranus", false},
			},
		},
		{
			"Who painted the Mona Lisa?",
			[]answerData{
				{"Leonardo da Vinci", true},
				{"Michelangelo", false},
				{"Raphael", false},
				{"Donatello", false},
			},
		},
		{
			"What is the capital of Australia?",
			[]answerData{
				{"Canberra", true},
				{"Sydney", false},
				{"Melbourne", false},
				{"Brisbane", false},
			},
		},
		{
			"How many bones are in the adult human body?",
			[]answerData{
				{"206", true},
				{"205", false},
				{"207", false},
				{"204", false},
			},
		},
		{
			"What is the smallest prime number?",
			[]answerData{
				{"2", true},
				{"1", false},
				{"3", false},
				{"0", false},
			},
		},
		{
			"Who invented the telephone?",
			[]answerData{
				{"Alexander Graham Bell", true},
				{"Thomas Edison", false},
				{"Nikola Tesla", false},
				{"Guglielmo Marconi", false},
			},
		},
		{
			"What is the hardest natural substance on Earth?",
			[]answerData{
				{"Diamond", true},
				{"Gold", false},
				{"Iron", false},
				{"Platinum", false},
			},
		},
		{
			"In which country would you find the ancient city of Petra?",
			[]answerData{
				{"Jordan", true},
				{"Egypt", false},
				{"Greece", false},
				{"Turkey", false},
			},
		},
	}

	for i, q := range questions {
		question, _ := createQuestion(q.text, 10, i+1, q.answers)
		gkQuiz.AddQuestion(*question)
	}

	r.quizzes[quizID.String()] = gkQuiz
}

// Helper types and functions

type answerData struct {
	text      string
	isCorrect bool
}

func createQuestion(text string, points int, position int, answersData []answerData) (*quiz.Question, error) {
	questionText, err := quiz.NewQuestionText(text)
	if err != nil {
		return nil, err
	}

	questionPoints, err := quiz.NewPoints(points)
	if err != nil {
		return nil, err
	}

	question, err := quiz.NewQuestion(quiz.NewQuestionID(), questionText, questionPoints, position)
	if err != nil {
		return nil, err
	}

	for i, answerData := range answersData {
		answerText, err := quiz.NewAnswerText(answerData.text)
		if err != nil {
			return nil, err
		}

		answer, err := quiz.NewAnswer(quiz.NewAnswerID(), answerText, answerData.isCorrect, i+1)
		if err != nil {
			return nil, err
		}

		question.AddAnswer(*answer)
	}

	return question, nil
}
