import React from 'react';

function Header() {
    return (
        <header className="flex justify-between items-center p-6 bg-gradient-to-r from-white to-blue-50 shadow-xl">
            <div className="logo">
                <h1 className="text-2xl font-bold text-blue-800">Luka</h1>
            </div>
            
            <nav className="navigation">
                <ul className="flex space-x-10">
                    <li><a className="text-blue-800 text-lg font-medium hover:text-blue-600 transition-colors duration-300">Home</a></li>
                    <li><a className="text-blue-800 text-lg font-medium hover:text-blue-600 transition-colors duration-300">About</a></li>
                    <li><a className="text-blue-800 text-lg font-medium hover:text-blue-600 transition-colors duration-300">How It Works</a></li>
                    <li><a className="text-blue-800 text-lg font-medium hover:text-blue-600 transition-colors duration-300">Features</a></li>
                    <li><a className="text-blue-800 text-lg font-medium hover:text-blue-600 transition-colors duration-300">FAQs</a></li>
                    <li><a className="text-blue-800 text-lg font-medium hover:text-blue-600 transition-colors duration-300">Contact</a></li>
                </ul>
            </nav>
            
            <div className="cta-buttons flex space-x-6">
                <button className="px-8 py-3 bg-blue-600 text-white font-semibold rounded-lg shadow-lg hover:bg-blue-700 focus:outline-none transition-all duration-300">Start Selling</button>
                <button className="px-8 py-3 bg-white text-blue-600 font-semibold border-2 border-blue-600 rounded-lg shadow-lg hover:bg-blue-50 focus:outline-none transition-all duration-300">Shop Now</button>
            </div>
        </header>
    );
}

export default Header;
